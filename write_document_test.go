package kdl

import (
	"bufio"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentWritesCorrectly(t *testing.T) {

	doc := NewDocument()

	n1 := NewNode("abc")
	n1.AddArg(NewNumberValue(big.NewFloat(2.0), ""))
	n1.AddArg(NewStringValue("foo", ""))

	n2 := NewNode("def")
	n2.SetProp("zoom", NewStringValue("voom", ""))
	n2.SetProp("quox", NewBoolValue(false, ""))

	n1.AddChild(n2)
	doc.AddNode(n1)

	n3 := NewNode("ghi jkl")
	doc.AddNode(n3)

	var s strings.Builder
	w := writer{writer: bufio.NewWriter(&s)}

	err := writeDocument(&w, &doc)
	assert.NoError(t, err)

	w.writer.Flush()
	assert.Equal(t, `abc 2 "foo" {
    def quox=false zoom="voom"
}
"ghi jkl"
`, s.String())
}
