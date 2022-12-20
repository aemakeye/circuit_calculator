package neo4j

import (
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/config"
	"github.com/aemakeye/circuit_calculator/internal/drawio"
	"github.com/aemakeye/circuit_calculator/internal/storage"
	"github.com/neo4j/neo4j-go-driver/v4/neo4j"
	"go.uber.org/zap"
	"io"
	"strings"
	"sync"
	"text/template"
)

const (
	relationQueueSize = 100
)

type Controller struct {
	Logger *zap.Logger
	Config neo4j.Config
	Driver *neo4j.Driver
}

var instance *Controller
var once sync.Once

func NewNeo4j(logger *zap.Logger, c *config.CConfig) (dbc *Controller, err error) {
	once.Do(func() {
		logger.Info("creating neo4j controlling structure")
		instance = &Controller{}
	})

	//TODO recall to close driver properly
	driver, err := neo4j.NewDriver("bolt://"+c.Neo4j.Endpoint,
		neo4j.BasicAuth(c.Neo4j.User, c.Neo4j.Password, ""))
	if err != nil {
		logger.Error("failed to create neo4j driver instance",
			zap.Error(err),
		)
		return &Controller{
			Logger: nil,
			Config: neo4j.Config{},
			Driver: nil,
		}, err
	}

	instance.Driver = &driver
	return instance, nil
}

// decision on element type is based on Class attribute value and made externally for this func
func nodeItemAdapter(item storage.Item) *NodeDTO {
	return &NodeDTO{
		UUID:     item.UUID,
		ID:       item.ID,
		Value:    item.Value,
		Class:    item.Class,
		SubClass: item.SubClass,
	}
}

// decision on element type is based on Class attribute value and made externally for this func
func relationItemAdapter(item storage.Item) *RelationDTO {
	return &RelationDTO{
		UUID:     item.UUID,
		ID:       item.ID,
		SourceId: item.SourceId,
		TargetId: item.TargetId,
		ExitX:    item.ExitX,
		ExitY:    item.ExitY,
		EntryX:   item.EntryX,
		EntryY:   item.EntryY,
	}
}

func (c *Controller) PushNode(logger *zap.Logger, dto *NodeDTO) (uuid string, id string, err error) {
	cypherq := "MERGE (item:Element {" +
		"uuid: '" + dto.UUID + "', " +
		"id: '" + fmt.Sprintf("%d", dto.ID) + "', " +
		"Value: '" + dto.Value + "', " +
		"Class: '" + dto.Class + "', " +
		"SubClass: '" + dto.SubClass + "'" +
		"}) " +
		"RETURN COALESCE(item.uuid,\"\")+':'+COALESCE(item.id,\"\")"
	driver := *c.Driver
	session := driver.NewSession(neo4j.SessionConfig{AccessMode: neo4j.AccessModeWrite})
	defer session.Close()

	tresult, err := session.WriteTransaction(
		func(tx neo4j.Transaction) (interface{}, error) {
			result, e := tx.Run(
				cypherq, map[string]interface{}{},
			)
			_ = result
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
		return "", "", err
	}
	logger.Info("Node in DB",
		zap.String("uuid:id", tresult.(string)),
	)
	tresultSplit := strings.Split(tresult.(string), ":")
	return tresultSplit[0], tresultSplit[1], nil
}

func (c *Controller) PushRelation(logger *zap.Logger, dto *RelationDTO) (uuid string, id string, err error) {
	//TODO: return (specific) error if no source or target for relation/edge
	//https://neo4j.com/docs/cypher-manual/current/clauses/merge/#merge-merge-on-a-relationship
	cypherqTemplate := template.New("pushRelation")
	cypherqTemplate, err = cypherqTemplate.Parse(`MATCH
		(source:Element {uuid: '{{.uuid'}, id: '{{.sourceId}}')})
		(target:Element {uuid: '{{.uuid'}, id: '{{.targetId}}')})
		MERGE (source) - [r:connected] - (target)
		RETURN r
		`,
	)
	if err != nil {
		logger.Error("error while creating relation between elements",
			zap.String("uuid", dto.UUID),
			zap.Int("source id", dto.SourceId),
			zap.Int("target id", dto.TargetId),
			zap.Error(err),
		)
		return "", "", err
	}
	return "", "", err
}

func (c *Controller) PushItem(logger *zap.Logger, item storage.Item) (string, string, error) {
	//TODO need to push all items atonce, because need to create all nodes first and edges after.
	logger.Info("pushing item",
		zap.String("UUID", item.UUID),
		zap.Int("ID", item.ID),
	)

	if item.Class == "lines" {
		uuid, id, err := c.PushRelation(logger, relationItemAdapter(item))
		if err != nil {
			logger.Error("failed to create relation",
				zap.String("UUID", item.UUID),
				zap.Int("ID", item.ID),
				zap.Error(err),
			)
			return "", "", err
		}
		return uuid, id, nil
	}
	_, cAllowed := drawio.ItemAvailableClass[item.Class]
	if item.Class != "lines" && cAllowed == true {
		uuid, id, err := c.PushNode(logger, nodeItemAdapter(item))
		if err != nil {
			logger.Error("failed to create node",
				zap.String("UUID", item.UUID),
				zap.Int("ID", item.ID),
				zap.Error(err),
			)
			return "", "", err
		}
		return uuid, id, nil
	}
	return "", "", fmt.Errorf("item is not a node and not an edge")
}

func (c *Controller) PushDiagram(logger *zap.Logger, diagram io.Reader) (uuid string, err error) {

	return uuid, nil
}

//func NewDTO(mx *MxCell, uuid string) (interface{}, error) {
//	//TODO: combine to single type of elements here, no need to split here
//	if shape, ok := mx.Style.attrs["shape"]; ok {
//		shapeNameArr := strings.Split(shape, ".")
//		shapeNameKind := shapeNameArr[len(shapeNameArr)-2]
//		el := EElementDTO{
//			UUID:  uuid,
//			ID:    mx.Id,
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
//			ID:       mx.Id,
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
