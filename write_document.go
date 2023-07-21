package kdl

import (
	"bufio"
	"bytes"
	"strings"

	"golang.org/x/exp/slices"
)

// writeArgs serializes Node's arguments.
func writeArgs(w *writer, n *Node) error {

	args := n.Args
	if len(args) == 0 {
		return nil
	}

	for i := range args {

		arg := &args[i]

		if err := writeValue(w, arg); err != nil {
			return err
		}

		// Join arguments with a single space
		if i+1 < len(args) {
			if err := writeSpace(w); err != nil {
				return err
			}
		}
	}

	return nil
}

// writeProps serializes [Node]'s properties.
func writeProps(w *writer, n *Node) error {

	p := n.Props
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

		// Join properties with a single space
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
		if err := writeArgs(w, n); err != nil {
			return err
		}
	}

	if len(n.Props) > 0 {
		if err := writeSpace(w); err != nil {
			return err
		}
		if err := writeProps(w, n); err != nil {
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
		if i+1 < len(nodes) {
			if err := w.writer.WriteByte('\n'); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *Document) WriteString() (string, error) {
	var buf bytes.Buffer
	w := writer{writer: bufio.NewWriter(&buf)}
	err := writeDocument(&w, d)
	if err != nil {
		return "", err
	}
	if err := w.writer.WriteByte('\n'); err != nil {
		return "", err
	}
	w.writer.Flush()
	return buf.String(), nil
}
