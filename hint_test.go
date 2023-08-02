package kdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHintStoresPresence(t *testing.T) {
	foo := Hint("foo")
	assert.True(t, foo.IsPresent())
	assert.False(t, foo.IsAbsent())
	id, ok := foo.Get()
	assert.Equal(t, Identifier("foo"), id)
	assert.True(t, ok)
	assert.NotPanics(t, func() { foo.MustGet() })
}

func TestEmptyHintStoresPresence(t *testing.T) {
	bar := Hint("")
	assert.True(t, bar.IsPresent())
	assert.False(t, bar.IsAbsent())
	id, ok := bar.Get()
	assert.Equal(t, Identifier(""), id)
	assert.True(t, ok)
	assert.NotPanics(t, func() { bar.MustGet() })
}

func TestNoHintStoresAbsence(t *testing.T) {
	baz := NoHint()
	assert.False(t, baz.IsPresent())
	assert.True(t, baz.IsAbsent())
	_, ok := baz.Get()
	assert.False(t, ok)
	assert.Panics(t, func() { baz.MustGet() })
}
