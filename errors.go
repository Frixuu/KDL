package kdl

import (
	"errors"
	"fmt"
	"io"
)

var (
	ErrDifferentKeys   = errors.New("all keys of KDLObject to convert to document should be the same")
	ErrInvalidKeyChar  = errors.New("invalid character for key")
	ErrInvalidSyntax   = errors.New("invalid syntax")
	ErrInvalidNumValue = fmt.Errorf("%w: bad numeric value", ErrInvalidSyntax)
	ErrInvalidTypeTag  = errors.New("value has invalid KDL type tag")
	ErrUnexpectedEOF   = io.ErrUnexpectedEOF

	errKeyOnly  = errors.New("internal only: key only")
	errEndOfObj = errors.New("internal only: end of KDLObject")
)

// ErrWithPosition wraps an error,
// adding information where in the document it happened.
type ErrWithPosition struct {
	Err    error // The original error.
	Line   int   // Line where the error occurred, 1-indexed.
	Column int   // Column where the error occurred, 0-indexed.
}

// Error formats an error message.
func (e *ErrWithPosition) Error() string {
	return fmt.Sprintf("%s\n(on line %d, column %d)", e.Err.Error(), e.Line, e.Column)
}

// Unwrap returns the original error.
func (e *ErrWithPosition) Unwrap() error {
	return e.Err
}

// addPosInfo wraps an error, adding position information from context.
func addPosInfo(err error, r *reader) error {
	return &ErrWithPosition{Err: err, Line: r.line, Column: r.pos}
}
