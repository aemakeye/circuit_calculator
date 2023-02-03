package drawio

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"go.uber.org/zap"
	"io"
	"strconv"
	"strings"
	"sync"
)

const (
	ItemClassLines      = "lines"
	ItemClassResistors  = "resistors"
	ItemClassCapacitors = "capacitors"
	ItemClassInductors  = "inductors"
)

var ItemAvailableClass = map[string]struct{}{
	ItemClassResistors:  {},
	ItemClassCapacitors: {},
	ItemClassInductors:  {},
	ItemClassLines:      {},
}

type Controller struct {
	logger *zap.Logger
}

type Item struct {
	UUID     string
	EID      int
	Value    string
	Class    string
	SubClass string
	SourceId int
	TargetId int
	ExitX    float32
	ExitY    float32
	EntryX   float32
	EntryY   float32
	Props    map[string]interface{}
	Error    error
}

var instance *Controller
var once sync.Once

func NewController(logger *zap.Logger) *Controller {
	once.Do(func() {
		logger.Info("creating drawio controller instance")
		instance = &Controller{logger: logger}
	})
	return instance
}

type Mxfile struct {
	//XMLName xml.Name `xml:"host,attr"`
	Diagram struct {
		Id string `xml:"id,attr"`
		//TODO: try "a>b>c" read.go 70 with branch
		MxGraphModel struct {
			Root struct {
				MxCells []MxCell `xml:"mxCell"`
			} `xml:"root"`
		} `xml:"mxGraphModel"`
	} `xml:"diagram"`
}

type MxCell struct {
	Id     int     `xml:"id,attr"`
	Style  style   `xml:"style,attr"`
	Value  string  `xml:"value,attr"`
	Source int     `xml:"source,attr,omitempty"`
	Target int     `xml:"target,attr,omitempty"`
	ExitX  float32 `xml:"exitX,attr,omitempty"`
	ExitY  float32 `xml:"exitY,attr,omitempty"`
	EntryX float32 `xml:"entryX,attr,omitempty"`
	EntryY float32 `xml:"entryY,attr,omitempty"`
}

type style struct {
	attrs map[string]string
}

func (sh *style) UnmarshalXMLAttr(attr xml.Attr) error {
	attrList := strings.Split(attr.Value, ";")
	attrMap := make(map[string]string)
	for i := range attrList {
		k, v := func(as string) (string, string) {
			x := strings.Split(as, "=")
			if len(x) == 2 {
				return x[0], x[1]
			} else {
				return "", ""
			}
		}(attrList[i])
		if k != "" {
			attrMap[k] = v
		}
	}
	sh.attrs = attrMap
	return nil
}

func NewItemDTO(mx *MxCell, uuid string) ItemDTO {
	item := ItemDTO{
		UUID:  uuid,
		ID:    mx.Id,
		Value: mx.Value,
	}

	if shape, ok := mx.Style.attrs["shape"]; ok {
		shapeNameArr := strings.Split(shape, ".")
		item.Class = shapeNameArr[len(shapeNameArr)-2]
		item.SubClass = shapeNameArr[len(shapeNameArr)-1]
	}

	if _, ok := mx.Style.attrs["endArrow"]; ok {
		// i believe line has exit/entry attributes when both source and target are set
		ExitX, _ := strconv.ParseFloat(mx.Style.attrs["exitX"], 32)
		ExitY, _ := strconv.ParseFloat(mx.Style.attrs["exitY"], 32)
		EntryX, _ := strconv.ParseFloat(mx.Style.attrs["entryX"], 32)
		EntryY, _ := strconv.ParseFloat(mx.Style.attrs["entryY"], 32)

		item.SourceId = mx.Source
		item.TargetId = mx.Target
		item.ExitX = float32(ExitX)
		item.ExitY = float32(ExitY)
		item.EntryX = float32(EntryX)
		item.EntryY = float32(EntryY)
		item.Class = "lines"
		item.SubClass = "line"
	}

	return item
}

// ReadInDiagram converts incoming document from xml to a channel of diagram.Item  objects
func (c *Controller) XmlToItems(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader, ch chan Item) (uuid string, err error) {
	logger.Info("processing new document")
	D := &Mxfile{}
	xmlbytes, err := io.ReadAll(xmldoc)
	if err != nil {
		logger.Error("could not read in the document",
			zap.Error(err),
		)
		return uuid, err
	}

	err = xml.Unmarshal(xmlbytes, D)
	if err != nil && err.Error() != "EOF" {
		logger.Error("can not unmarshal document",
			zap.Error(err),
		)
		return uuid, err
	}

	uuid = D.Diagram.Id
	if uuid == "" {
		return uuid, fmt.Errorf("no diagram id in document")
	}
	for _, item := range D.Diagram.MxGraphModel.Root.MxCells {
		if item.Style.attrs == nil {
			logger.Debug("skipping element with no attributes",
				zap.Int("id", item.Id),
			)
			continue
		}
		di := NewItemDTO(&item, uuid)
		ch <- ItemsAdapter(di)
	}

	return uuid, err
}

func ItemsAdapter(item ItemDTO) Item {
	return Item{
		UUID:     item.UUID,
		EID:      item.ID,
		Value:    item.Value,
		Class:    item.Class,
		SubClass: item.SubClass,
		SourceId: item.SourceId,
		TargetId: item.TargetId,
		ExitX:    item.ExitX,
		ExitY:    item.ExitY,
		EntryX:   item.EntryY,
		EntryY:   item.EntryY,
	}
}
