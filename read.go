package kdl

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"unicode"
)

func readQuotedString(r *reader) (s string, err error) {

	s, err = readQuotedStringInner(r)
	if err != nil {
		return
	}

	s = strings.ReplaceAll(s, "\\/", "/")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s, err = strconv.Unquote(`"` + s + `"`)
	return
}

func readQuotedStringInner(r *reader) (string, error) {

	count := 1

	start, err := r.readRune()
	if err != nil {
		return "", err
	}

	if start != '"' {
		return "", ErrInvalidSyntax
	}

	for {

		bytes, err := r.peekN(count)
		if err != nil {
			r.discard(count)
			return string(bytes), err
		}

		ch := rune(bytes[len(bytes)-1])
		if ch == '\\' {

			bs, err := r.peekN(count + 1)
			if err != nil {
				r.discard(count)
				return string(bytes), err
			}

			next := bs[len(bs)-1] == byte('"')
			if next {
				count += 2
				continue
			}

		} else if ch == '"' {

			toRet := string(bytes[:len(bytes)-1])

			/*
				temp, err := r.peekN(count + 1)
				if err != nil {
					if err.Error() != eof {
						return toRet, err
					}
					r.discard(count)
					return toRet, nil
				}


				ch = rune(temp[len(temp)-1])
				if !unicode.IsSpace(ch) && ch != ';' {
					return toRet, ErrInvalidSyntax
				}
			*/

			r.discard(count)
			return toRet, nil
		}

		count++
	}
}

func readRawString(r *reader) (string, error) {

	leadingPoundCount := 0
	length := 0

	for {

		length++

		bytes, err := r.peekN(length)
		if err != nil {
			return "", err
		}

		ch := rune(bytes[len(bytes)-1])
		if ch == '#' {
			leadingPoundCount++
			continue
		} else if ch == '"' {
			break
		} else {
			return "", ErrInvalidSyntax
		}
	}

	start := length
	length++
	closingPoundCount := 0
	dqStart := false

	for {

		bytes, err := r.peekN(length)
		if err != nil {
			return "", err
		}

		ch := rune(bytes[len(bytes)-1])
		if ch == '"' {
			dqStart = true
			length++
			continue
		}

		if dqStart && ch == '#' {
			closingPoundCount++
		} else {
			closingPoundCount = 0
			dqStart = false
		}

		if closingPoundCount == leadingPoundCount {
			r.discard(length)
			return string(bytes[start : len(bytes)-leadingPoundCount-1]), nil
		}

		length++
	}
}

var trueBytes = []byte{'t', 'r', 'u', 'e'}
var falseBytes = []byte{'f', 'a', 'l', 's', 'e'}
var ErrExpectedBool = fmt.Errorf("%w: expected boolean", ErrInvalidSyntax)

func readBool(r *reader) (bool, error) {

	var expected []byte

	start, err := r.peek()
	if err != nil {
		return false, err
	}

	switch start {
	case 't':
		expected = trueBytes
	case 'f':
		expected = falseBytes
	default:
		return false, ErrInvalidSyntax
	}

	next, err := r.isNext(expected)
	if err != nil {
		return false, err
	}

	if next {
		return start == 't', nil
	}

	return false, ErrExpectedBool
}

var nullBytes = []byte{'n', 'u', 'l', 'l'}
var ErrExpectedNull = fmt.Errorf("%w: expected null", ErrInvalidSyntax)

func readNull(r *reader) error {

	next, err := r.isNext(nullBytes)
	if err != nil {
		return err
	}

	if next {
		return nil
	}

	return ErrExpectedNull
}

func readNumber(r *reader) (*big.Float, error) {

	length := 0
	dotCount := 0

	for {

		length++

		bytes, err := r.peekN(length)
		if err != nil && err.Error() != eof {
			return big.NewFloat(0), err
		}

		ch := rune(bytes[len(bytes)-1])
		if ch == dot {
			dotCount++
			if dotCount > 1 {
				return big.NewFloat(0), ErrInvalidNumValue
			}
		}

		if ch == ';' || unicode.IsSpace(ch) ||
			ch == '/' || (err != nil && err.Error() == eof) {
			rawStr := string(bytes[0 : len(bytes)-1])
			if err != nil && err.Error() == eof {
				rawStr = string(bytes)
			}

			r.discard(length - 1)

			str := strings.ReplaceAll(rawStr, "_", "")
			value, err := strconv.ParseFloat(str, 64)
			if err == nil {
				return big.NewFloat(value), nil
			}

			val, err := strconv.ParseInt(str, 0, 64)
			return new(big.Float).SetInt64(val), err
		}
	}
}
