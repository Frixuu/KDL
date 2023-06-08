package kdl

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

const (
	eof = "EOF"

	asterisk   = '*'
	backslash  = '\\'
	dash       = '-'
	dot        = '.'
	dquote     = '"'
	equals     = '='
	newline    = '\n'
	pound      = '#'
	semicolon  = ';'
	slash      = '/'
	space      = ' '
	underscore = '_'

	openBracket      = '{'
	closeBracket     = '}'
	openParenthesis  = '('
	closeParenthesis = ')'
)

func parseObjects(r *reader, hasOpen bool, key string) (KDLObjects, error) {
	var t KDLObjects
	var objects []KDLObject
	for {
		obj, err := parseObject(r)
		if err == nil {
			if obj != nil {
				objects = append(objects, obj)
			}
		} else if err.Error() == eof || errors.Is(err, errEndOfObj) {
			if obj != nil {
				objects = append(objects, obj)
			}
			return NewKDLObjects(key, objects), nil
		} else {
			return t, addPosInfo(err, r)
		}
	}
}

func parseObject(kdlr *reader) (KDLObject, error) {

	r, err := kdlr.peek()
	if err != nil {
		return nil, err
	}

	if err = readUntilSignificant(kdlr); err != nil {
		return nil, err
	}

	if r == closeBracket {
		kdlr.discard(1)
		return nil, errEndOfObj
	}

	skipNext, _ := kdlr.isNext([]byte{slash, dash})
	if skipNext {
		parseKey(kdlr)
	}

	key, err := parseKey(kdlr)

	if err != nil {
		if errors.Is(err, errKeyOnly) {
			return NewKDLDefault(key), nil
		}
		return nil, err
	}

	var objects []KDLObject
	for {

		if err = readUntilSignificant(kdlr); err != nil {
			return nil, err
		}

		r, err := kdlr.readRune()
		if err != nil && err.Error() != eof {
			return nil, err
		}

		if r == backslash {
			peek, err := kdlr.peek()
			if err == nil && peek == newline {
				kdlr.discard(1)
				continue
			}
		}

		if r == newline || r == semicolon ||
			(err != nil && err.Error() == eof) {
			if len(objects) == 0 {
				return NewKDLDefault(key), nil
			} else if len(objects) == 1 {
				return objects[0], nil
			} else {
				return ConvertToDocument(objects)
			}
		} else if unicode.IsSpace(r) {
			continue
		}

		kdlr.unreadRune()
		skipNext, _ := kdlr.isNext([]byte{slash, dash})
		if skipNext {
			r, err = kdlr.peek()
			if err != nil {
				if err.Error() == eof {
					return ConvertToDocument(objects)
				}
				return nil, err
			}
		}

		if err = readUntilSignificant(kdlr); err != nil {
			if errors.Is(err, io.EOF) {
				return ConvertToDocument(objects)
			}
			return nil, err
		}

		obj, err := parseVal(kdlr, key, r)
		if err != nil {
			if errors.Is(err, errEndOfObj) {
				return ConvertToDocument(objects)
			}
			return nil, err
		}
		if !skipNext {
			objects = append(objects, obj)
		}
	}
}

func parseKey(kdlr *reader) (string, error) {
	var key strings.Builder
	isQuoted := false
	prev := newline

	for {
		r, err := kdlr.readRune()
		if err != nil {
			if err.Error() == eof {
				return tryUnquote(key), errKeyOnly
			}
			return key.String(), err
		}

		if (!isQuoted && unicode.IsSpace(r)) || r == newline ||
			((unicode.IsSpace(r) || r == equals) && prev == dquote) {
			if key.Len() < 1 {
				continue
			} else if r == newline {
				return tryUnquote(key), errKeyOnly
			} else {
				return tryUnquote(key), nil
			}
		}

		invalid :=
			(key.Len() < 1 && unicode.IsNumber(r)) ||
				(!isQuoted && unicode.IsSpace(r)) || r == equals
		if invalid {
			return key.String(), ErrInvalidKeyChar
		}

		if key.Len() < 1 {
			isQuoted = r == dquote
		}
		if prev == backslash && r == backslash {
			prev = newline
		} else if prev == backslash && r == dquote {
			prev = newline
		} else {
			prev = r
		}
		key.WriteRune(r)
	}
}

func parseVal(kdlr *reader, key string, r rune) (KDLObject, error) {
	value, err := parseValue(kdlr, key, r)
	if err == nil || errors.Is(err, ErrInvalidNumValue) {
		return value, err
	}

	if errors.Is(err, errEndOfObj) {
		return value, err
	}

	node, err := parseKey(kdlr)

	if err != nil && !errors.Is(err, ErrInvalidKeyChar) {
		if errors.Is(err, errKeyOnly) {
			return NewKDLObjects(key, []KDLObject{NewKDLDefault(node)}), nil
		}
		return nil, err
	}

	if kdlr.lastRead() != equals {
		return NewKDLObjects(key, []KDLObject{NewKDLDefault(node)}), nil
	}
	r, err = kdlr.peek()
	if err != nil {
		return nil, err
	}

	obj, err := parseValue(kdlr, node, r)
	if err != nil {
		return nil, err
	}

	return NewKDLObjects(key, []KDLObject{obj}), nil
}

func parseValue(r *reader, key string, ch rune) (KDLObject, error) {

	for ch == ' ' {
		var err error
		ch, err = r.readRune()
		if err != nil {
			return nil, err
		}
	}

	if unicode.IsNumber(ch) || ch == '-' {
		n, err := readNumber(r)
		return KDLNumber{key: key, value: NewNumberValue(n, "")}, err
	}

	switch ch {
	case '"':
		s, err := readQuotedString(r)
		return KDLString{key: key, value: NewStringValue(s, "")}, err
	case 'n':
		return KDLNull{key: key, value: NewNullValue()}, readNull(r)
	case 't', 'f':
		b, err := readBool(r)
		return KDLBool{key: key, value: NewBoolValue(b, "")}, err
	case 'r':
		r.discard(1)
		s, err := readRawString(r)
		return KDLRawString{key: key, value: NewStringValue(s, "")}, err
	case '{':
		r.discard(1)
		return parseObjects(r, true, key)
	case '}':
		return nil, errEndOfObj
	}

	return nil, ErrInvalidSyntax
}
