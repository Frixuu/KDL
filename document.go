package kdl

type Document struct {
	Nodes []Node
}

func NewDocument() Document {
	return Document{Nodes: make([]Node, 0)}
}
