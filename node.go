package kdl

// Node is an object in a KDL Document.
type Node struct {
	TypeHint TypeHint             // Optional hint about the type of this node.
	Name     Identifier           // Name of the node.
	Args     []Value              // Ordered arguments of the node.
	Props    map[Identifier]Value // Unordered properties of the node. CAN BE NIL.
	Children []Node               // Ordered children of the node.
}

// NewNode creates a new KDL node.
func NewNode(name string) Node {
	return Node{
		Name: Identifier(name),
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

// GetProp returns a property of this Node.
func (n *Node) GetProp(key Identifier) Value {
	props := n.Props
	if props == nil {
		return newInvalidValue()
	}
	return props[key]
}

// SetProp sets or replaces a property of this Node.
func (n *Node) SetProp(key Identifier, value Value) {
	if n.Props == nil {
		n.Props = make(map[Identifier]Value)
	}
	n.Props[key] = value
}
