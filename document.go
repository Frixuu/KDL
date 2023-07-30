package kdl

// Document is a top-level unit of the KDL format.
type Document struct {
	Nodes []Node
}

// NewDocument creates a new Document.
func NewDocument() Document {
	nodes := make([]Node, 0)
	return Document{Nodes: nodes}
}

// nodeParent is an object that can have many children nodes, ie. Document or a Node.
type nodeParent interface {
	AddChild(n Node)
}

// AddChild adds a node to this Document.
func (d *Document) AddChild(n Node) {
	d.Nodes = append(d.Nodes, n)
}
