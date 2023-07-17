package kdl

import (
	"bufio"
	"math/big"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentWritesCorrectly(t *testing.T) {

	var s strings.Builder
	w := writer{writer: bufio.NewWriter(&s)}

	err := writeDocument(&w, &Document{
		Nodes: []Node{
			{
				Name: "abc",
				Args: []Value{
					NewNumberValue(big.NewFloat(2.0), ""),
					NewStringValue("foo", ""),
				},
				Children: []Node{
					{
						Name: "def",
						Props: map[Identifier]Value{
							"zoom": NewStringValue("voom", ""),
							"quox": NewBoolValue(false, ""),
						},
					},
				},
			},
			{
				Name: "ghi jkl",
			},
		},
	})

	w.writer.Flush()
	assert.NoError(t, err)
	assert.Equal(t, `abc 2 "foo" {
    def quox=false zoom="voom"
}
"ghi jkl"
`, s.String())
}
