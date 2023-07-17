package kdl

import (
	"errors"
	"fmt"
	"io"
	"unicode"
)

func readNodes(r *reader) ([]Node, error) {
	r.depth++
	// TODO
	r.depth--
	return []Node{}, nil
}

func readNode(r *reader) (Node, error) {
	node := NewNode("")

	ch, err := r.peek()
	if ch == '(' && err != nil {
		hint, err := readTypeHint(r)
		if err != nil {
			return node, err
		}
		node.TypeHint = hint
	}

	name, err := readIdentifier(r)
	if err != nil {
		return node, err
	}

	node.Name = name

	for {

		err = readUntilSignificant(r)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return node, nil
			}
			return node, err
		}

		ch, err := r.peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				return node, nil
			}
			return node, err
		}

		if isNewLine(ch) {
			r.discard(1)
			return node, nil
		} else if ch == ';' {
			r.discard(1)
			return node, nil
		} else if ch == '}' {
			return node, nil
		} else if ch == '{' {
			r.discard(1)
			children, err := readNodes(r)
			if err != nil {
				return node, err
			}
			for i := range children {
				node.AddChild(children[i])
			}
		} else {
			err = readArgOrProp(r, &node)
			if err != nil {
				return node, err
			}
		}
	}
}

var ErrInvalidPropertyKey = fmt.Errorf("%w: bad property key type", ErrInvalidSyntax)
var ErrPropertyKeyIsNumber = fmt.Errorf("%w (cannot be a number)", ErrInvalidPropertyKey)

// readArgOrProp reads an argument or a property
// and adds them to the provided Node definition.
func readArgOrProp(r *reader, dest *Node) error {

	// A "slashdash" comment silences the whole argument or property
	slashdash, err := r.isNext(charsSlashDash[:])
	if err != nil {
		return err
	}
	if slashdash {
		r.discard(len(charsSlashDash))
	}

	ch, err := r.peek()
	if err != nil {
		return err
	}

	// Values (NOT property keys) can be prepended by an additional type hint
	typeHint, err := readTypeHint(r)
	if err != nil {
		return err
	}

	if unicode.IsDigit(ch) {
		num, err := readNumber(r)
		if err != nil {
			return err
		}
		ch, err = r.peek()
		if err == nil || errors.Is(err, io.EOF) || ch == ';' || ch == '}' || isWhitespace(ch) {
			if !slashdash {
				dest.AddArg(NewNumberValue(num, typeHint))
			}
			return nil
		} else if ch == '=' {
			return ErrPropertyKeyIsNumber
		} else {
			return ErrInvalidSyntax
		}
	}

	length := 1
	for {
		name, err := r.peekN(length)
		if err != nil {
			if errors.Is(err, io.EOF) {
				name, err = r.peekN(length - 1)
				if err != nil {
					return nil
				}
				arg := string(name)
				r.discard(length - 1)
				if !slashdash {
					dest.AddArg(NewStringValue(arg, ""))
				}
				return nil
			}
			return err
		}

		length++
	}
}

func readIdentifier(r *reader) (Identifier, error) {
	ch, err := r.peek()
	if err != nil {
		return "", err
	}

	if ch == '"' {
		s, err := readQuotedString(r)
		return Identifier(s), err
	}

	// TODO read bare identifier

	return "", ErrInvalidSyntax
}

// skipUntilNewLine discards the reader to the next new line character.
//
// If afterBreak is true, the reader is positioned after the newline break.
// If it is false, the reader is positioned just before a newline rune. (singular, in case of CRLF)
func skipUntilNewLine(r *reader, afterBreak bool) error {

	for {

		// CRLF is a special case as it spans two runes, so we check it first
		if isCrlf, err := r.isNext(charsCRLF[:]); isCrlf && err == nil {
			if afterBreak {
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

		if isNewLine(ch) {
			if afterBreak {
				r.discard(1)
			}
			break
		}

		r.discard(1)
	}

	return nil
}

// readUntilSignificant allows the provided reader to skip whitespace and comments.
//
// Note: this method will NOT skip over new lines.
func readUntilSignificant(r *reader) error {

outer:
	for {

		ch, err := r.peek()
		if err != nil {
			return err
		}

		if isWhitespace(ch) {
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
