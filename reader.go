package kdl

import (
	"bufio"
	"bytes"
)

type reader struct {
	reader *bufio.Reader
	line   int
	pos    int
	depth  int
}

func wrapReader(r *bufio.Reader) *reader {
	return &reader{reader: r, line: 1, pos: 0}
}

func (r *reader) readRune() (rune, error) {
	ch, _, err := r.reader.ReadRune()
	if ch == '\n' {
		r.line++
		r.pos = 0
	} else {
		r.pos++
	}

	return ch, err
}

func (r *reader) discardRunes(count int) {
	for i := 0; i < count; i++ {
		_, _ = r.readRune()
	}
}

func (r *reader) discardBytes(count int) {
	s, _ := r.peekBytes(count)
	for _, b := range s {
		var nl byte = '\n'
		if b == nl {
			r.line++
			r.pos = 0
		} else {
			r.pos++
		}
	}
	r.reader.Discard(count)
}

// peekBytes tries to return next N bytes without advancing the reader.
func (r *reader) peekBytes(count int) ([]byte, error) {
	return r.reader.Peek(count)
}

func (r *reader) peekRune() (rune, error) {
	ch, _, err := r.reader.ReadRune()
	if err != nil {
		return ch, err
	}

	err = r.reader.UnreadRune()
	return ch, err
}

func (r *reader) isNext(expected []byte) (bool, error) {

	next, err := r.peekBytes(len(expected))
	if err != nil {
		return false, err
	}

	return bytes.Equal(next, expected), nil
}
