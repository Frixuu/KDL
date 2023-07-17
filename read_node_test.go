package kdl

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadsSimpleNode(t *testing.T) {
	reader := readerFromString("foo bar\n(name)baz quox")

	n, err := readNode(reader)
	assert.NoError(t, err)
	assert.Equal(t, Node{
		Name: "foo",
		Args: []Value{NewStringValue("bar", "")},
	}, n)

	n, err = readNode(reader)
	assert.NoError(t, err)
	assert.Equal(t, Node{
		Name:     "baz",
		TypeHint: "name",
		Args:     []Value{NewStringValue("quox", "")},
	}, n)

	_, err = readNode(reader)
	assert.ErrorIs(t, err, io.EOF)
}
