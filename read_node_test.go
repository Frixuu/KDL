package kdl

import (
	"io"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadsSimpleNode(t *testing.T) {

	reader := readerFromString("foo \"bar\" (abc)2\n(name)baz \"quox\"")

	n, err := readNode(&reader)
	assert.NoError(t, err)
	assert.Equal(t, Node{
		Name:  "foo",
		Props: map[Identifier]Value{},
		Args: []Value{
			NewStringValue("bar", NoHint()),
			NewIntegerValue(big.NewInt(2), Hint("abc")),
		},
	}, n)

	// previous readNode consumes the \n
	n, err = readNode(&reader)
	assert.NoError(t, err)
	assert.Equal(t, Node{
		Name:     "baz",
		TypeHint: Hint("name"),
		Props:    map[Identifier]Value{},
		Args:     []Value{NewStringValue("quox", NoHint())},
	}, n)

	_, err = readNode(&reader)
	assert.ErrorIs(t, err, io.EOF)
}

func TestReadsNodeWithChildren(t *testing.T) {

	reader := readerFromString(`repo type="git" {
	/-mirror "foo"
	mirror "bar"; mirror "baz"
}`)

	n, err := readNode(&reader)
	assert.NoError(t, err)
	assert.Equal(t, "git", n.Props["type"].RawValue)
	assert.Equal(t, 2, len(n.Children))
	assert.Equal(t, "baz", n.Children[1].Args[0].AsString())
}

func TestReadsLineContinuation(t *testing.T) {
	reader := readerFromString("\"foo\" \\\n\"bar\"")
	n, err := readNode(&reader)
	assert.NoError(t, err)
	assert.EqualValues(t, "foo", n.Name)
	assert.Equal(t, 1, len(n.Args))
	assert.Equal(t, "bar", n.Args[0].AsString())
}

func TestReadsMultilineCommentAfterLineContinuation(t *testing.T) {
	reader := readerFromString(`"foo" \ /*

	*/
	"bar" \ /*

*/
"baz"`)
	n, err := readNode(&reader)
	assert.NoError(t, err)
	assert.EqualValues(t, "foo", n.Name)
	assert.Equal(t, 2, len(n.Args))
	assert.Equal(t, "bar", n.Args[0].AsString())
	assert.Equal(t, "baz", n.Args[1].AsString())
}

func TestReadsRTL(t *testing.T) {
	input := `الطاب الطاب=1 الطاب=2`
	reader := readerFromString(input)
	n, err := readNode(&reader)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(n.Props))
	assert.EqualValues(t, 2, n.Props["الطاب"].AsInteger().Int64())
}
