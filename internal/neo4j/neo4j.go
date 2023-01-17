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

type pushResult struct {
	uuid  string
	id    int
	error error
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

func (c *Controller) PushNodes(logger *zap.Logger, chitem <-chan drawio.Item, pr chan pushResult) {
	driver := c.Driver
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	for item := range chitem {
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
			pr <- pushResult{
				uuid:  item.UUID,
				id:    item.EID,
				error: err,
			}
			continue
		}
		logger.Info("Node in DB",
			zap.String("uuid:id", tresult.(string)),
		)
		//tresultSplit := strings.Split(tresult.(string), ":")
		//id, err := strconv.Atoi(tresultSplit[1])
		pr <- pushResult{
			uuid:  item.UUID,
			id:    item.EID,
			error: nil,
		}

	}
}

func (c *Controller) PushRelation(logger *zap.Logger, chitem <-chan drawio.Item, pr chan pushResult) {
	driver := c.Driver
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	var tbuf []byte
	twrbuf := bytes.NewBuffer(tbuf)
	//TODO: return (specific) error if no source or target for relation/edge
	//https://neo4j.com/docs/cypher-manual/current/clauses/merge/#merge-merge-on-a-relationship
	for item := range chitem {
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
			pr <- pushResult{
				uuid:  item.UUID,
				id:    0,
				error: err,
			}
			continue
		}
		logger.Info("relation pushed",
			zap.Ints("Source id: %d, Target id: %d",
				[]int{tresult.(drawio.Item).SourceId, tresult.(drawio.Item).TargetId}),
		)
		//tresultSplit := strings.Split(tresult.(string), ":")
		//id, err := strconv.Atoi(tresultSplit[1])
		pr <- pushResult{
			uuid:  item.UUID,
			id:    item.EID,
			error: nil,
		}
	}
}

// TODO: decom this
//func (c *Controller) PushItem(logger *zap.Logger, item storage.Item) (string, string, error) {
//	//TODO need to push all items atonce, because need to create all nodes first and edges after.
//	logger.Info("pushing item",
//		zap.String("UUID", item.UUID),
//		zap.Int("EID", item.EID),
//	)
//
//	if item.Class == "lines" {
//		uuid, id, err := c.PushRelation(logger, relationItemAdapter(item))
//		if err != nil {
//			logger.Error("failed to create relation",
//				zap.String("UUID", item.UUID),
//				zap.Int("EID", item.EID),
//				zap.Error(err),
//			)
//			return "", "", err
//		}
//		return uuid, id, nil
//	}
//	_, cAllowed := drawio.ItemAvailableClass[item.Class]
//	if item.Class != "lines" && cAllowed == true {
//		uuid, id, err := c.PushNode(logger, nodeItemAdapter(item))
//		if err != nil {
//			logger.Error("failed to create node",
//				zap.String("UUID", item.UUID),
//				zap.Int("EID", item.EID),
//				zap.Error(err),
//			)
//			return "", "", err
//		}
//		return uuid, id, nil
//	}
//	return "", "", fmt.Errorf("item is not a node and not an edge")
//}

// PushItems reads data from channel, pushes Nodes first, and then relations
func (c *Controller) PushItems(logger *zap.Logger, items <-chan drawio.Item) (err error) {
	itemsParsed := 0
	relChanQueue := make(chan drawio.Item, relationQueueSize)

	nodeChan := make(chan drawio.Item)

	for range items {
		select {
		case item := <-items:
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
				logger.Debug("Sending relation to wait queue",
					zap.String("uuid", item.UUID),
					zap.Int("id", item.EID),
				)
			}

			itemsParsed++
			logger.Debug("item parsed",
				zap.String("uuid", item.UUID),
				zap.Int("id", item.EID),
			)
		default:

		}
	}

	return nil
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

//func NewDTO(mx *MxCell, uuid string) (interface{}, error) {
//	//TODO: combine to single type of elements here, no need to split here
//	if shape, ok := mx.Style.attrs["shape"]; ok {
//		shapeNameArr := strings.Split(shape, ".")
//		shapeNameKind := shapeNameArr[len(shapeNameArr)-2]
//		el := EElementDTO{
//			UUID:  uuid,
//			EID:    mx.Id,
//			Value: mx.Value,
//			Kind:  shapeNameKind,
//			Type:  shapeNameArr[len(shapeNameArr)-1],
//		}
//		if kind, kok := KnownEEKinds[el.Kind]; !kok {
//			return nil, fmt.Errorf("unsupported kind of element %s (id %d)", kind, mx.Id)
//		}
//		return el, nil
//	}
//	// if mx is a line
//	if _, ok := mx.Style.attrs["endArrow"]; ok {
//
//		if mx.Source == 0 {
//			return nil, fmt.Errorf("no source in line attributes (id: %d)", mx.Id)
//		}
//
//		if mx.Target == 0 {
//			return nil, fmt.Errorf("no target in line attributes (id: %d)", mx.Id)
//		}
//
//		// i believe line has exit/entry attributes when both source and target are set
//		ExitX, _ := strconv.ParseFloat(mx.Style.attrs["exitX"], 32)
//		ExitY, _ := strconv.ParseFloat(mx.Style.attrs["exitY"], 32)
//		EntryX, _ := strconv.ParseFloat(mx.Style.attrs["entryX"], 32)
//		EntryY, _ := strconv.ParseFloat(mx.Style.attrs["entryY"], 32)
//
//		el := LineDTO{
//			UUID:     uuid,
//			EID:       mx.Id,
//			SourceId: mx.Source,
//			TargetId: mx.Target,
//			ExitX:    float32(ExitX),
//			ExitY:    float32(ExitY),
//			EntryX:   float32(EntryX),
//			EntryY:   float32(EntryY),
//		}
//		return el, nil
//	}
//	return nil, fmt.Errorf("unknown element (id: %d)", mx.Id)
//}
