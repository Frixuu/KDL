package kdl

import (
	"errors"
	"unsafe"
)

var (
	emptyData  []byte     = make([]byte, 1)
	emptyPtr   *byte      = unsafe.SliceData(emptyData)
	emptyIdent Identifier = Identifier(unsafe.String(emptyPtr, 0))
)

// TypeHint is an optional Identifier associated with a Value.
type TypeHint struct {
	hint Identifier
}

// Hint constructs a present TypeHint.
func Hint(name string) TypeHint {
	if len(name) > 0 {
		return TypeHint{hint: Identifier(name)}
	}
	return TypeHint{hint: emptyIdent}
}

// NoHint constructs a missing TypeHint.
func NoHint() TypeHint {
	return TypeHint{}
}

// IsPresent returns true if the hint exists.
func (h TypeHint) IsPresent() bool {
	return len(h.hint) > 0 || (unsafe.StringData(string(h.hint)) == emptyPtr)
}

// IsAbsent returns true if the hint does not exist.
func (h TypeHint) IsAbsent() bool {
	return len(h.hint) == 0 && (unsafe.StringData(string(h.hint)) != emptyPtr)
}

// Get returns the inner type hint, if it exists.
func (h TypeHint) Get() (Identifier, bool) {
	return h.hint, h.IsPresent()
}

var errTypeHintAbsent = errors.New("called MustGet on an absent kdl.TypeHint")

// MustGet returns the inner type hint or panics, if it does not exist.
func (h TypeHint) MustGet() Identifier {
	if h.IsAbsent() {
		panic(errTypeHintAbsent)
	}
	return h.hint
}
