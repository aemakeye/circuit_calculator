package drawio

import (
	"bytes"
	"context"
	"encoding/xml"
	"fmt"
	"github.com/aemakeye/circuit_calculator/internal/calculator"
	"go.uber.org/zap"
	"io"
	"strconv"
	"strings"
	"sync"
)

var ItemAvailableClass = map[string]struct{}{
	"resistors":  struct{}{},
	"capacitors": {},
	"inductors":  {},
	"lines":      {},
}

type Controller struct {
	logger *zap.Logger
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

func NewItemDTO(mx *MxCell, uuid string) *ItemDTO {
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

	return &item
}

// ReadInDiagram converts incoming document from xml to a channel of calculator.Item  objects
func (c *Controller) ReadInDiagram(ctx context.Context, logger *zap.Logger, xmldoc *bytes.Reader) (uuid string, _ <-chan calculator.Item, err error) {
	ch := make(chan calculator.Item, 1)
	defer close(ch)

	logger.Info("processing new document")
	D := &Mxfile{}
	xmlbytes, err := io.ReadAll(xmldoc)
	if err != nil {
		logger.Error("could not read in the document",
			zap.Error(err),
		)
		return uuid, nil, err
	}

	err = xml.Unmarshal(xmlbytes, D)
	if err != nil && err.Error() != "EOF" {
		logger.Error("can not unmarshal document",
			zap.Error(err),
		)
		return uuid, nil, err
	}

	uuid = D.Diagram.Id
	if uuid == "" {
		return uuid, nil, fmt.Errorf("no diagram id in document")
	}
	for _, item := range D.Diagram.MxGraphModel.Root.MxCells {
		ch <- ItemsAdapter(c.logger, *NewItemDTO(&item, uuid))
	}
	return uuid, ch, err
}

func ItemsAdapter(logger *zap.Logger, item ItemDTO) (citems calculator.Item) {
	return calculator.Item{
		UUID:     item.UUID,
		ID:       item.ID,
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

func (c *Controller) UpdateDiagram(ctx context.Context, logger *zap.Logger, diaUUID string) error {
	panic("implement me")

}
