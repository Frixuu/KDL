package kdl

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"golang.org/x/exp/maps"
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
	keys := maps.Keys(p)
	slices.Sort(keys)

	for i, key := range keys {

		value := p[key]

		if err := writeIdentifier(w, key); err != nil {
			return err
		}
		if err := w.writer.WriteByte('='); err != nil {
			return err
		}
		if err := writeValue(w, &value); err != nil {
			return err
		}

		// Join properties with a single space
		if i+1 < len(keys) {
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

	if err := writeTypeHint(w, n.TypeHint); err != nil {
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

// Write writes the Document to an io.Writer.
func (d *Document) Write(w io.Writer) error {
	bw := writer{writer: bufio.NewWriter(w)}
	if err := writeDocument(&bw, d); err != nil {
		return err
	}
	if err := bw.writer.WriteByte('\n'); err != nil {
		return err
	}
	return bw.writer.Flush()
}

// WriteString marshals the Document to a new string.
func (d *Document) WriteString() (string, error) {
	var buf bytes.Buffer
	err := d.Write(&buf)
	return buf.String(), err
}
