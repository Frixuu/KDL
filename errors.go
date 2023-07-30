package kdl

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
)

var (
	ErrInvalidSyntax       = errors.New("invalid syntax")
	ErrInvalidNumValue     = fmt.Errorf("%w: bad numeric value", ErrInvalidSyntax)
	ErrUnexpectedSlashdash = fmt.Errorf("%w: unexpected slashdash", ErrInvalidSyntax)

	ErrUnexpectedEOF   = io.ErrUnexpectedEOF
	ErrInvalidEncoding = errors.New("document is not UTF-8 encoded")
)

// ErrWithPosition wraps an error,
// adding information where in the document did it happen.
type ErrWithPosition struct {
	Err    error // The original error.
	Line   int   // Line where the error occurred, 1-indexed.
	Column int   // Column where the error occurred, 0-indexed.
}

// Error formats an error message.
func (e *ErrWithPosition) Error() string {
	var s strings.Builder
	innerMsg := e.Err.Error()
	s.Grow(len(innerMsg) + 30)
	s.WriteString(innerMsg)
	s.WriteString(" [line ")
	s.WriteString(strconv.Itoa(e.Line))
	s.WriteString(", column ")
	s.WriteString(strconv.Itoa(e.Column))
	s.WriteString("]")
	return s.String()
}

// Unwrap returns the original error.
func (e *ErrWithPosition) Unwrap() error {
	return e.Err
}

// addErrPosInfo wraps an error, adding position information from context.
func addErrPosInfo(err error, r *reader) error {
	return &ErrWithPosition{Err: err, Line: r.line, Column: r.pos}
}
