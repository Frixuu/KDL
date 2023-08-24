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

// AddArg adds an element as an order-sensitive argument of this Node.
func (n *Node) AddArg(arg interface{}) error {
	v, err := ValueOf(arg)
	if err != nil {
		return err
	}
	n.AddArgValue(v)
	return nil
}

// AddArgValue adds a Value as an order-sensitive argument of this Node.
func (n *Node) AddArgValue(arg Value) {
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

// HasProp returns true if this Node has a property of that name.
func (n *Node) HasProp(key Identifier) bool {
	props := n.Props
	if props == nil {
		return false
	}
	val, ok := props[key]
	if !ok {
		return false
	}
	return val.Type != TypeInvalid
}

// SetProp sets or replaces a property of this Node.
func (n *Node) SetProp(key Identifier, value interface{}) error {

	v, err := ValueOf(value)
	if err != nil {
		return err
	}

	n.SetPropValue(key, v)
	return nil
}

// SetPropValue sets or replaces a property of this Node.
func (n *Node) SetPropValue(key Identifier, value Value) {
	props := n.Props
	if props != nil {
		props[key] = value
	} else {
		n.Props = map[Identifier]Value{key: value}
	}
}

// RemoveProp removes a property from this Node.
func (n *Node) RemoveProp(key Identifier) {
	props := n.Props
	if props == nil {
		return
	}
	delete(props, key)
}
