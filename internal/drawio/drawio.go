package drawio

import (
	"encoding/xml"
	"fmt"
	"go.uber.org/zap"
	"strconv"
	"strings"
)

var KnownEEKinds = map[string]struct{}{
	"resistors":  struct{}{},
	"capacitors": {},
	"inductors":  {},
}

type DiagramBuilder struct {
	Logger *zap.Logger
	Mxfile *Mxfile
}

type Mxfile struct {
	//XMLName xml.Name `xml:"host,attr"`
	Diagram struct {
		Id           string `xml:"id,attr"`
		MxGraphModel struct {
			Root struct {
				MxCells []MxCell `xml:"mxCell"`
			} `xml:"root"`
		} `xml:"mxGraphModel"`
	} `xml:"diagram"`
}

type MxCell struct {
	Id     int    `xml:"id,attr"`
	Style  style  `xml:"style,attr"`
	Value  string `xml:"value,attr"`
	Source int    `xml:"source,attr,omitempty"`
	Target int    `xml:"target,attr,omitempty"`
}

type style struct {
	attrs map[string]string
}

func (sh *style) UnmarshalXMLAttr(attr xml.Attr) error {
	attrList := strings.Split(attr.Value, ";")
	//*sh = style{shape: "", style: attr.Value}
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

func NewDTO(mx *MxCell, uuid string) (interface{}, error) {
	if shape, ok := mx.Style.attrs["shape"]; ok {
		shapeNameArr := strings.Split(shape, ".")
		shapeNameKind := shapeNameArr[len(shapeNameArr)-2]
		el := EElementDTO{
			UUID:  uuid,
			ID:    mx.Id,
			Value: mx.Value,
			Kind:  shapeNameKind,
			Type:  shapeNameArr[len(shapeNameArr)-1],
		}
		if kind, kok := KnownEEKinds[el.Kind]; !kok {
			return nil, fmt.Errorf("unsupported kind of element %s (id %d)", kind, mx.Id)
		}
		return el, nil
	}
	// if mx is a line
	if _, ok := mx.Style.attrs["endArrow"]; ok {

		if mx.Source == 0 {
			return nil, fmt.Errorf("no source in line attributes (id: %d)", mx.Id)
		}

		if mx.Target == 0 {
			return nil, fmt.Errorf("no target in line attributes (id: %d)", mx.Id)
		}

		// i believe line has exit/entry attributes when both source and target are set
		ExitX, _ := strconv.ParseFloat(mx.Style.attrs["exitX"], 32)
		ExitY, _ := strconv.ParseFloat(mx.Style.attrs["exitY"], 32)
		EntryX, _ := strconv.ParseFloat(mx.Style.attrs["entryX"], 32)
		EntryY, _ := strconv.ParseFloat(mx.Style.attrs["entryY"], 32)

		el := Line{
			UUID:     uuid,
			ID:       mx.Id,
			SourceId: mx.Source,
			TargetId: mx.Target,
			ExitX:    float32(ExitX),
			ExitY:    float32(ExitY),
			EntryX:   float32(EntryX),
			EntryY:   float32(EntryY),
		}
		return el, nil
	}
	return nil, fmt.Errorf("unknown element (id: %d)", mx.Id)
}
