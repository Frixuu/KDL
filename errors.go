package kdl

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrEmptyArray      = errors.New("array is empty")
	ErrDifferentKeys   = errors.New("all keys of KDLObject to convert to document should be the same")
	ErrInvalidKeyChar  = errors.New("invalid character for key")
	ErrInvalidNumValue = errors.New("invalid numeric value")
	ErrInvalidSyntax   = errors.New("invalid syntax")
	ErrInvalidTypeTag  = errors.New("value has invalid KDL type tag")
	ErrUnexpectedEOF   = io.ErrUnexpectedEOF

	errKeyOnly     = errors.New("internal only: key only")
	errEndOfObj    = errors.New("internal only: end of KDLObject")
	errNothingLeft = errors.New("internal only: nothing else left to parse")
)

// ErrWithPosition wraps an error,
// adding information where in the document it happened.
type ErrWithPosition struct {
	Err    error // The original error.
	Line   int
	Column int
}

func (e *ErrWithPosition) Error() string {
	return fmt.Sprintf("%s\n(on line %d, column %d)", e.Err.Error(), e.Line, e.Column)
}

func (e *ErrWithPosition) Unwrap() error {
	return e.Err
}

func addPosInfo(err error, r *reader) error {
	return &ErrWithPosition{Err: err, Line: r.line, Column: r.pos}
}
