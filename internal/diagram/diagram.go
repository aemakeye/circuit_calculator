package diagram

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
