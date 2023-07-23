package kdl

import "github.com/samber/mo"

// Node is an object in a KDL Document.
type Node struct {
	Name     Identifier
	Args     []Value
	Props    map[Identifier]Value
	Children []Node
	TypeHint mo.Option[Identifier]
}

// NewNode creates a new KDL node.
func NewNode(name string) Node {
	return Node{
		Name:  Identifier(name),
		Props: make(map[Identifier]Value),
	}
}

// AddArg adds a Value as an order-sensitive argument of this Node.
func (n *Node) AddArg(arg Value) {
	n.Args = append(n.Args, arg)
}

// AddChild adds another Node as an order-sensitive child of this Node.
func (n *Node) AddChild(child Node) {
	n.Children = append(n.Children, child)
}

// SetProp sets or replaces a property of this Node.
func (n *Node) SetProp(key Identifier, value Value) {
	n.Props[key] = value
}
