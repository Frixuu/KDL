package kdl

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProp(t *testing.T) {
	n := NewNode("foo")
	assert.False(t, n.HasProp("bar"))
	n.SetProp("baz", NewIntegerValue(big.NewInt(1), NoHint()))
	assert.False(t, n.HasProp("bar"))
	n.SetProp("bar", NewIntegerValue(big.NewInt(1), NoHint()))
	assert.True(t, n.HasProp("bar"))
	n.RemoveProp("bar")
	assert.False(t, n.HasProp("bar"))
}
