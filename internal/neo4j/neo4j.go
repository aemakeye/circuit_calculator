package neo4j

import (
	"bytes"
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/drawio"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j/dbtype"
	"go.uber.org/zap"
	"sync"
	"text/template"
)

const (
	relationQueueSize = 100
	schemaUUID        = "uuid"
	schemaID          = "eid"
	schemaValue       = "Value"
	schemaClass       = "Class"
	schemaSubClass    = "SubClass"
)

type Controller struct {
	Logger   *zap.Logger
	Driver   neo4j.Driver
	user     string
	password string
	url      string
}

var instance *Controller
var once sync.Once

func NewController(logger *zap.Logger, url string, user string, password string) (dbc *Controller, err error) {
	once.Do(func() {
		logger.Info("creating neo4j controlling structure")
		instance = &Controller{}
	})

	//TODO recall to close driver properly
	driver, err := neo4j.NewDriver("bolt://"+url,
		neo4j.BasicAuth(user, password, ""))
	if err != nil {
		logger.Error("failed to create neo4j driver instance",
			zap.Error(err),
		)
		return &Controller{
			Logger:   logger,
			Driver:   driver,
			user:     user,
			password: password,
			url:      url,
		}, err
	}

	instance.Driver = driver
	return instance, nil
}

func (c *Controller) PushNodes(logger *zap.Logger, chitem <-chan drawio.Item, pr chan drawio.Item, noMoreNodes chan struct{}) {
	driver := c.Driver
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		err := session.Close()
		if err != nil {
			logger.Error("Failed to close neo4j session")
		} else {
			logger.Debug("Closing neo4j Session")
		}
	}()

	for {
		select {
		case item, ok := <-chitem:
			if ok {
				cypherq := "MERGE (item:Element {" +
					schemaUUID + ": '" + item.UUID + "', " +
					schemaID + ": '" + fmt.Sprintf("%d", item.EID) + "', " +
					schemaValue + ": '" + item.Value + "', " +
					schemaClass + ": '" + item.Class + "', " +
					schemaSubClass + ": '" + item.SubClass + "'" +
					"}) " +
					"RETURN COALESCE(item.uuid,\"\")+':'+COALESCE(item.id,\"\")"

				tresult, err := session.WriteTransaction(
					func(tx neo4j.Transaction) (interface{}, error) {
						result, e := tx.Run(
							cypherq, map[string]interface{}{},
						)
						if e != nil {
							return nil, e
						}
						if result.Next() {
							return result.Record().Values[0], nil
						}
						return nil, nil
					},
				)

				if err != nil {
					logger.Error("error on transaction",
						zap.Error(err),
					)
					item.Error = err
					pr <- item
					continue
				}
				logger.Info("Node in DB",
					zap.String("uuid:id", tresult.(string)),
				)

				pr <- item
			}
		case <-noMoreNodes:
			logger.Info("parsed all nodes")
			return

		default:
		}
	}
}

func (c *Controller) PushRelations(logger *zap.Logger, chitem <-chan drawio.Item, pr chan drawio.Item, noMoreRels chan struct{}) {
	driver := c.Driver
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer func() {
		err := session.Close()
		if err != nil {
			logger.Error("Failed to close neo4j session")
		} else {
			logger.Debug("Closing neo4j Session")
		}
	}()

	var tbuf []byte
	twrbuf := bytes.NewBuffer(tbuf)
	//TODO: return (specific) error if no source or target for relation/edge
	//https://neo4j.com/docs/cypher-manual/current/clauses/merge/#merge-merge-on-a-relationship
	for {
		select {
		case item, ok := <-chitem:
			if ok {
				twrbuf.Reset()
				cypherqTemplate := template.New("pushRelation")
				cypherqTemplate, err := cypherqTemplate.Parse(`MATCH
		(source:Element {` + schemaUUID + `: '{{.UUID}}', ` + schemaID + `: '{{.SourceId}}' }),
		(target:Element {` + schemaUUID + `: '{{.UUID}}', ` + schemaID + `: '{{.TargetId}}' })
		MERGE (source) - [r:connected] - (target)
		RETURN r
		`)
				err = cypherqTemplate.Execute(twrbuf, item)
				if err != nil {
					logger.Error("failed to render cypher query template",
						zap.Error(err),
					)
					continue
				}
				tresult, err := session.WriteTransaction(
					func(tx neo4j.Transaction) (interface{}, error) {
						result, e := tx.Run(
							twrbuf.String(), map[string]interface{}{},
						)
						if e != nil {
							return nil, e
						}
						//in case relation created, dbtype.Relationship is returned.
						//let us convert it to drawio.Item ant return
						if result.Next() {
							r := result.Record().Values[0]
							return drawio.Item{
								UUID:     item.UUID,
								EID:      int(r.(dbtype.Relationship).Id),
								Value:    "",
								Class:    drawio.ItemClassLines,
								SubClass: "",
								SourceId: int(r.(dbtype.Relationship).StartId),
								TargetId: int(r.(dbtype.Relationship).EndId),
								ExitX:    0,
								ExitY:    0,
								EntryX:   0,
								EntryY:   0,
								Props:    r.(dbtype.Relationship).Props,
							}, nil
						}
						logger.Error("neo4j transaction failed, empty return value for relation ",
							zap.Ints("SourceID, TargetID", []int{item.SourceId, item.TargetId}),
						)
						return nil, fmt.Errorf("neo4j transaction failed, empty return value for relation")
					},
				)
				if err != nil {
					logger.Error("error while creating relation between elements",
						zap.String("uuid", item.UUID),
						zap.Int("source id", item.SourceId),
						zap.Int("target id", item.TargetId),
						zap.Error(err),
					)
					item.Error = err
					pr <- item
					continue
				}
				logger.Info("relation pushed",
					zap.Ints("Source id: %d, Target id: %d",
						[]int{tresult.(drawio.Item).SourceId, tresult.(drawio.Item).TargetId}),
				)
				pr <- item
			}
		case <-noMoreRels:
			return

		default:

		}
	}
}

// PushItems reads data from channel, pushes Nodes first, and then Relations
func (c *Controller) PushItems(logger *zap.Logger, items <-chan drawio.Item, pr chan drawio.Item, noMoreItems chan struct{}) {
	defer close(pr)
	relChanQueue := make(chan drawio.Item, relationQueueSize)
	relChanQueueItems := 0
	relChan := make(chan drawio.Item)
	nodeChan := make(chan drawio.Item)
	noMoreNodes := make(chan struct{})
	noMoreRels := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		c.PushNodes(logger, nodeChan, pr, noMoreNodes)
		wg.Done()
	}()

	go func() {
		c.PushRelations(logger, relChan, pr, noMoreRels)
		wg.Done()
	}()

OuterLoop:
	for {
		select {
		case item, ok := <-items:
			if ok {
				node, err := IsNode(&item)
				if err != nil {
					logger.Error("error",
						zap.Error(err),
					)
					continue
				}
				if node {
					nodeChan <- item
					logger.Debug("Pushing node",
						zap.String("uuid", item.UUID),
						zap.Int("id", item.EID),
					)
				} else {
					relChanQueue <- item
					relChanQueueItems++
					logger.Debug("Sending relation to wait queue",
						zap.String("uuid", item.UUID),
						zap.Int("id", item.EID),
					)
				}
				logger.Debug("item parsed",
					zap.String("uuid", item.UUID),
					zap.Int("id", item.EID),
				)
			}
		case <-noMoreItems:
			noMoreNodes <- struct{}{}
			break OuterLoop

		default:
		}
	}

	for i := 0; i < relChanQueueItems; i++ {
		iq := <-relChanQueue
		relChan <- iq
	}
	noMoreRels <- struct{}{}
	wg.Wait()
	return
}

func IsNode(item *drawio.Item) (bool, error) {
	_, ok := drawio.ItemAvailableClass[item.Class]
	if !ok {
		return false, fmt.Errorf("item class %s is not supported", item.Class)
	}
	isNode := item.Class != drawio.ItemClassLines
	if ok && isNode {
		return true, nil
	} else {
		return false, nil
	}
}
