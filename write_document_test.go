package kdl

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDocumentWritesCorrectly(t *testing.T) {

	s, err := (&Document{
		Nodes: []Node{
			{
				Name: "abc",
				Args: []Value{
					NewFloatValue(big.NewFloat(2.0), noHint),
					NewStringValue("foo", noHint),
				},
				Children: []Node{
					{
						Name: "def",
						Props: map[Identifier]Value{
							"zoom": NewStringValue("voom", noHint),
							"quox": NewBoolValue(false, noHint),
						},
					},
				},
			},
			{
				Name: "ghi jkl",
			},
		},
	}).WriteString()

	assert.NoError(t, err)
	assert.Equal(t, `abc 2.0 "foo" {
    def quox=false zoom="voom"
}
"ghi jkl"
`, s)
}
