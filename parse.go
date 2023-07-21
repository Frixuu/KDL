package kdl

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
)

//go:generate go run tools/generate_test_cases/generate.go

func ParseBufReader(br *bufio.Reader) (Document, error) {
	doc := NewDocument()
	r := wrapReader(br)

	for {

		_, err := r.peekRune()
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

func ParseReader(r io.Reader) (Document, error) {
	br := bufio.NewReader(r)
	return ParseBufReader(br)
}

func ParseBytes(b []byte) (Document, error) {
	bb := bytes.NewBuffer(b)
	return ParseReader(bb)
}

func ParseString(s string) (Document, error) {
	sb := bytes.NewBufferString(s)
	return ParseReader(sb)
}

func ParseFile(path string) (Document, error) {
	f, err := os.Open(path)
	if err != nil {
		return NewDocument(), err
	}
	defer f.Close()
	br := bufio.NewReader(f)
	return ParseBufReader(br)
}
