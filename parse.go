package kdl

import (
	"bufio"
	"bytes"
	"errors"
	"io"
)

func ParseDocument(data []byte) (Document, error) {
	doc := NewDocument()
	r := wrapReader(bufio.NewReader(bytes.NewBuffer(data)))

	for {

		_, err := r.peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return doc, err
		}

		node, err := readNode(r)
		if err != nil {
			return doc, err
		}

		doc.AddNode(node)
	}

	return doc, nil
}

func readNode(r *reader) (Node, error) {
	node := NewNode("")
	return node, nil
}

func skipWhitespace(r *reader) error {
	return nil
}
