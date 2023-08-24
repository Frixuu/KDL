package kdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProp(t *testing.T) {
	n := NewNode("foo")
	assert.False(t, n.HasProp("bar"))
	n.SetProp("baz", 1)
	assert.False(t, n.HasProp("bar"))
	n.SetProp("bar", 1)
	assert.True(t, n.HasProp("bar"))
	n.RemoveProp("bar")
	assert.False(t, n.HasProp("bar"))
}
