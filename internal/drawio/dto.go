package drawio

type NodeDTO struct {
	UUID        string
	ID          int
	Connections []int
}

type ResistorDTO struct {
	UUID        string
	ID          int
	Connections []int
	Text        []string
	Value       string
}

type CapacitorDTO struct {
	UUID        string
	ID          int
	Connections []int
	Text        []string
	Value       string
}

type InductanceDTO struct {
	UUID        string
	ID          int
	Connections []int
	Text        []string
	Value       string
}
