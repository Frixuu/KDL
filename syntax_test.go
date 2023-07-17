package kdl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowsValidBareIdentifiers(t *testing.T) {
	assert.True(t, isAllowedBareIdentifier("foo"))
	assert.True(t, isAllowedBareIdentifier("k"))
	assert.True(t, isAllowedBareIdentifier("a99999"))
	assert.True(t, isAllowedBareIdentifier("-bar"))
}

func TestDisallowsInvalidBareIdentifiers(t *testing.T) {
	assert.False(t, isAllowedBareIdentifier("foo bar"))
	assert.False(t, isAllowedBareIdentifier("1337"))
	assert.False(t, isAllowedBareIdentifier(`"quox"`))
	assert.False(t, isAllowedBareIdentifier("true"))
	assert.False(t, isAllowedBareIdentifier(""))
}
