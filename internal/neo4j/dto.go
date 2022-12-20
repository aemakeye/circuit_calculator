package neo4j

type NodeDTO struct {
	UUID     string
	ID       int
	Value    string
	Class    string
	SubClass string
}

type RelationDTO struct {
	UUID     string
	ID       int
	SourceId int
	TargetId int
	ExitX    float32
	ExitY    float32
	EntryX   float32
	EntryY   float32
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
