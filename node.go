package kdl

type Node struct {
	Name     Identifier
	Args     []Value
	Props    map[Identifier]Value
	Children []Node
	TypeHint Identifier
}

func NewNode(name string) Node {
	return Node{
		Name:     Identifier(name),
		Args:     make([]Value, 0),
		Props:    make(map[Identifier]Value),
		Children: make([]Node, 0),
		TypeHint: Identifier(""),
	}
}
