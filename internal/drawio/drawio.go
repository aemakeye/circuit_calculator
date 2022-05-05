package drawio

import (
	"encoding/xml"
	"go.uber.org/zap"
	"strings"
)

type DiagramBuilder struct {
	Logger *zap.Logger
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
	Shape  shape  `xml:"style,attr"`
	Source string `xml:"source,attr"`
	Target string `xml:"target,attr"`
}

type shape struct {
	shape string
	style string
}

func (sh *shape) UnmarshalXMLAttr(attr xml.Attr) error {
	attrList := strings.Split(attr.Value, ";")
	*sh = shape{shape: "", style: attr.Value}
	for a := range attrList {
		s := strings.Split(attrList[a], "=")
		if s[0] == "shape" {
			sh.shape = s[1]
		}
	}
	return nil
}
