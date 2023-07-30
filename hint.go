package kdl

import "errors"

// TypeHint is an optional Identifier associated with a Value.
type TypeHint struct {
	hint    Identifier
	present bool
}

// Hint constructs a present TypeHint.
func Hint(name string) TypeHint {
	return TypeHint{
		hint:    Identifier(name),
		present: true,
	}
}

// NoHint constructs a missing TypeHint.
func NoHint() TypeHint {
	return TypeHint{present: false}
}

// IsPresent returns true if the hint exists.
func (h TypeHint) IsPresent() bool {
	return h.present
}

// IsAbsent returns true if the hint does not exist.
func (h TypeHint) IsAbsent() bool {
	return !h.present
}

// Get returns the inner type hint, if it exists.
func (h TypeHint) Get() (Identifier, bool) {
	return h.hint, h.present
}

var errTypeHintAbsent = errors.New("called MustGet on an absent kdl.TypeHint")

// MustGet returns the inner type hint or panics, if it does not exist.
func (h TypeHint) MustGet() Identifier {
	if !h.present {
		panic(errTypeHintAbsent)
	}
	return h.hint
}
