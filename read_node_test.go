package kdl

import (
	"io"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadsSimpleNode(t *testing.T) {

	reader := readerFromString("foo \"bar\" (abc)2\n(name)baz \"quox\"")

	n, err := readNode(reader)
	assert.NoError(t, err)
	assert.Equal(t, Node{
		Name:  "foo",
		Props: map[Identifier]Value{},
		Args: []Value{
			NewStringValue("bar", noHint),
			NewIntegerValue(big.NewInt(2), hint("abc")),
		},
	}, n)

	// previous readNode consumes the \n
	n, err = readNode(reader)
	assert.NoError(t, err)
	assert.Equal(t, Node{
		Name:     "baz",
		TypeHint: hint("name"),
		Props:    map[Identifier]Value{},
		Args:     []Value{NewStringValue("quox", noHint)},
	}, n)

	_, err = readNode(reader)
	assert.ErrorIs(t, err, io.EOF)
}

func TestReadsNodeWithChildren(t *testing.T) {

	reader := readerFromString(`repo type="git" {
	/-mirror "foo"
	mirror "bar"; mirror "baz"
}`)

	n, err := readNode(reader)
	assert.NoError(t, err)
	assert.Equal(t, "git", n.Props["type"].RawValue)
	assert.Equal(t, 2, len(n.Children))
	assert.Equal(t, "baz", n.Children[1].Args[0].AsString())
}
