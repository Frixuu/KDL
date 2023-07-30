package kdl

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"strings"
)

//go:generate go run internal/tools/generate_test_cases/generate.go

func ParseBufReader(br *bufio.Reader) (Document, error) {
	doc := NewDocument()
	r := wrapReader(br)

	nodes, err := readNodes(&r)
	if err != nil {
		return doc, addErrPosInfo(err, &r)
	}

	doc.Nodes = nodes
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
	sr := strings.NewReader(s)
	wr := bufio.NewReaderSize(sr, len(s))
	return ParseBufReader(wr)
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
