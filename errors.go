package kdl

import (
	"errors"
	"io"
	"strconv"
	"strings"
)

var (
	// ErrInvalidSyntax is a base error for when
	// a parser comes across a document that is not spec-compliant.
	ErrInvalidSyntax = errors.New("invalid syntax")
	// ErrInvalidEncoding is a base error for when
	// an invalid UTF8 byte sequence is encountered.
	ErrInvalidEncoding = errors.New("document is not UTF-8 encoded")
	// ErrUnexpectedEOF is a base error for when
	// the data abruptly ends e.g. inside a string.
	ErrUnexpectedEOF = io.ErrUnexpectedEOF
)

// ErrWithPosition wraps an error,
// adding information where in the document did it occur.
type ErrWithPosition struct {
	Err    error // The original error.
	Line   int   // Line where the error occurred, 1-indexed.
	Column int   // Column where the error occurred, 0-indexed.
}

// Error formats an error message.
func (e *ErrWithPosition) Error() string {

	innerMsg := "null"
	err := e.Err
	if err != nil {
		innerMsg = err.Error()
	}

	var s strings.Builder
	s.Grow(len(innerMsg) + 24)
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
