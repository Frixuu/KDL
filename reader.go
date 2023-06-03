package kdlgo

import (
	"bufio"
	"bytes"
)

type reader struct {
	reader  *bufio.Reader
	line    int
	pos     int
	current rune
}

func newKDLReader(r *bufio.Reader) *reader {
	return &reader{line: 1, pos: 0, reader: r}
}

func (kdlr *reader) readRune() (rune, error) {
	r, _, err := kdlr.reader.ReadRune()
	if r == '\n' {
		kdlr.line++
		kdlr.pos = 0
	} else {
		kdlr.pos++
	}

	if err == nil {
		kdlr.current = r
	}

	return r, err
}

func (kdlr *reader) lastRead() rune {
	return kdlr.current
}

func (kdlr *reader) discardLine() error {
	_, err := kdlr.reader.ReadString('\n')
	if err != nil {
		return err
	}

	err = kdlr.reader.UnreadByte()
	return err
}

func (kdlr *reader) discard(count int) {
	s, _ := kdlr.peekX(count)
	for _, b := range s {
		var nl byte = '\n'
		if b == nl {
			kdlr.line++
			kdlr.pos = 0
		} else {
			kdlr.pos++
		}
	}
	kdlr.reader.Discard(count)
}

func (kdlr *reader) peekX(count int) ([]byte, error) {
	return kdlr.reader.Peek(count)
}

func (kdlr *reader) peek() (rune, error) {
	r, _, err := kdlr.reader.ReadRune()
	if err != nil {
		return r, err
	}

	err = kdlr.reader.UnreadRune()
	return r, err
}

func (kdlr *reader) unreadRune() error {
	err := kdlr.reader.UnreadRune()
	if err != nil {
		return err
	}

	peek, _ := kdlr.reader.Peek(1)
	var b byte = '\n'
	if peek[0] == b {
		kdlr.line--
	} else {
		kdlr.pos--
	}

	return nil
}

func (kdlr *reader) isNext(charset []byte) (bool, error) {
	peek, err := kdlr.peekX(len(charset))
	if err != nil {
		return false, err
	}

	if bytes.Compare(peek, charset) == 0 {
		kdlr.discard(len(charset))
		return true, nil
	}

	return false, nil
}
