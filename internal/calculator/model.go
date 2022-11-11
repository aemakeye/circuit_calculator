package calculator

type Diagram struct {
	UUID     string
	Body     string
	IsValid  bool
	Error    string
	Items    []Item
	Name     string
	Versions []DiagramVersion
}

type DiagramVersion struct {
	Version  string
	metadata string
}

type Item struct {
	UUID     string
	ID       int
	Value    string
	Class    string
	SubClass string
	SourceId int
	TargetId int
	ExitX    float32
	ExitY    float32
	EntryX   float32
	EntryY   float32
}
