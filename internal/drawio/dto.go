package drawio

type EElementDTO struct {
	UUID  string
	ID    int
	Value string
	Kind  string
	Type  string
}

type Line struct {
	UUID     string
	ID       int
	SourceId int
	TargetId int
	ExitX    float32
	ExitY    float32
	EntryX   float32
	EntryY   float32
}
