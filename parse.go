package kdl

import (
	"bufio"
	"bytes"
	"errors"
	"io"

	"golang.org/x/exp/slices"
)

func ParseDocument(data []byte) (Document, error) {
	doc := NewDocument()
	r := wrapReader(bufio.NewReader(bytes.NewBuffer(data)))

	for {

		_, err := r.peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return doc, err
		}

		node, err := readNode(r)
		if err != nil {
			return doc, err
		}

		doc.AddNode(node)
	}

	return doc, nil
}

func readNode(r *reader) (Node, error) {
	node := NewNode("")
	return node, nil
}

var charsSlashDash = [...]byte{'/', '-'}

func readArgOrProp(r *reader, dest *Node) error {

	slashdash, err := r.isNext(charsSlashDash[:])
	if err != nil {
		return err
	}

	if slashdash {
		r.discard(2)
	}

	return nil
}

var whitespaceChars = [...]rune{
	0x9, 0x20, 0xa0, 0x1680,
	0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200a,
	0x202f, 0x205f,
	0x3000,
}

var charsStartComment = [...]byte{'/', '/'}
var charsStartCommentBlock = [...]byte{'/', '*'}
var charsEndCommentBlock = [...]byte{'*', '/'}
var charsCRLF = [...]byte{'\r', '\n'}
var charsNewLine = [...]rune{'\r', '\n', 0x85, 0xc, 0x2028, 0x2029}

// skipUntilNewLine discards the reader to the next new line character.
func skipUntilNewLine(r *reader, past bool) error {
outer:
	for {

		// CRLF is a special case as it spans two runes, so we check it first
		if isCrlf, err := r.isNext(charsCRLF[:]); isCrlf && err == nil {
			if past {
				r.discard(2)
			} else {
				// Leave the LF only to simplify later checks
				r.discard(1)
			}
			break
		}

		ch, err := r.peek()
		if err != nil {
			return err
		}

		for _, newLine := range charsNewLine {
			if ch == newLine {
				if past {
					r.discard(1)
				}
				break outer
			}
		}

		r.discard(1)
	}
	return nil
}

// readUntilSignificant allows the provided reader to skip whitespace and comments.
func readUntilSignificant(r *reader) error {

outer:
	for {

		ch, err := r.peek()
		if err != nil {
			return err
		}

		// Check the next rune for regular whitespace
		if slices.Contains(whitespaceChars[:], ch) {
			r.discard(1)
			continue
		}

		// Check for line continuation
		if ch == '\\' {
			r.discard(1)
			if err := skipUntilNewLine(r, true); err != nil {
				return err
			}
			continue
		}

		// Check for single-line comments
		if comment, err := r.isNext(charsStartComment[:]); comment && err == nil {
			r.discard(2)
			return skipUntilNewLine(r, false)
		}

		// Check for multiline comments
		if comment, err := r.isNext(charsStartCommentBlock[:]); comment && err == nil {
			r.discard(2)
			// Per spec, multiline comments can be nested, so we can't do naive ReadString("*/")
			depth := 1
		inner:
			for {

				start, err := r.isNext(charsStartCommentBlock[:])
				if err != nil {
					return err
				}

				if start {
					depth += 1
					r.discard(2)
					continue inner
				}

				end, err := r.isNext(charsEndCommentBlock[:])
				if err != nil {
					return err
				}

				if end {
					r.discard(2)
					depth -= 1
					if depth <= 0 {
						continue outer
					} else {
						continue inner
					}
				}

				r.discard(1)
			}
		}

		return nil
	}
}
