package kdl

// Document is a top-level unit of the KDL format.
type Document struct {
	Nodes []Node
}

// NewDocument creates a new Document.
func NewDocument() Document {
	return Document{Nodes: make([]Node, 0)}
}

// AddNode adds a node to this Document.
func (d *Document) AddNode(n Node) {
	d.Nodes = append(d.Nodes, n)
}
