package kdl

import (
	"strings"

	"golang.org/x/exp/slices"
)

func writeArgs(w *writer, a []Value) error {

	for i := range a {
		arg := &a[i]
		if err := writeValue(w, arg); err != nil {
			return err
		}
		if i+1 < len(a) {
			if err := writeSpace(w); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeProps(w *writer, p map[Identifier]Value) error {

	// If there are no props, there is nothing to write
	if len(p) == 0 {
		return nil
	}

	// Sort properties alphabetically
	props := toPairs(p)
	slices.SortFunc(props, func(a, b pair[Identifier, Value]) bool {
		return strings.Compare(string(a.key), string(b.key)) < 0
	})

	for i := range props {
		prop := &props[i]
		if err := writeIdentifier(w, prop.key); err != nil {
			return err
		}
		if err := w.writer.WriteByte('='); err != nil {
			return err
		}
		if err := writeValue(w, &prop.value); err != nil {
			return err
		}
		if i+1 < len(p) {
			if err := writeSpace(w); err != nil {
				return err
			}
		}
	}

	return nil
}

func writeNode(w *writer, n *Node) error {

	indent := strings.Repeat("    ", w.depth)
	if _, err := w.writer.WriteString(indent); err != nil {
		return err
	}

	if err := writeIdentifier(w, n.Name); err != nil {
		return err
	}

	if len(n.Args) > 0 {
		if err := writeSpace(w); err != nil {
			return err
		}
		if err := writeArgs(w, n.Args); err != nil {
			return err
		}
	}

	if len(n.Props) > 0 {
		if err := writeSpace(w); err != nil {
			return err
		}
		if err := writeProps(w, n.Props); err != nil {
			return err
		}
	}

	if len(n.Children) > 0 {

		if _, err := w.writer.WriteString(" {"); err != nil {
			return err
		}

		w.depth++
		for i := range n.Children {
			if err := w.writer.WriteByte('\n'); err != nil {
				return err
			}
			child := &n.Children[i]
			if err := writeNode(w, child); err != nil {
				return err
			}
		}
		w.depth--

		if err := w.writer.WriteByte('\n'); err != nil {
			return err
		}

		if _, err := w.writer.WriteString(indent); err != nil {
			return err
		}

		if err := w.writer.WriteByte('}'); err != nil {
			return err
		}
	}

	return nil
}

func writeDocument(w *writer, d *Document) error {

	nodes := d.Nodes

	for i := range nodes {
		node := &nodes[i]
		if err := writeNode(w, node); err != nil {
			return err
		}
		if err := w.writer.WriteByte('\n'); err != nil {
			return err
		}
	}

	return nil
}
