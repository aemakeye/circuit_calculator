package calculator

import "github.com/aemakeye/circuit_calculator/internal/storage"

type Diagram struct {
	UUID     string
	Body     string
	Error    string
	Items    []storage.Item
	Name     string
	Versions []DiagramVersion
}

type DiagramVersion struct {
	Version  string
	Metadata string
}
