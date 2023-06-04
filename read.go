package kdl

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
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

var (
	patternDecimal = regexp.MustCompile(`^[-+]?[0-9][_0-9]*(\.[0-9][_0-9]*)?([eE][-+]?[0-9][_0-9]*)?$`)
	patternHex     = regexp.MustCompile(`^[-+]?0x[0-9a-fA-F][_0-9a-fA-F]*$`)
	patternOctal   = regexp.MustCompile(`^[-+]?0o[0-7][_0-7]*$`)
	patternBinary  = regexp.MustCompile(`^[-+]?0b[01][_01]*$`)

	ErrBadDecimal = fmt.Errorf("%w (decimal does not match pattern)", ErrInvalidNumValue)
	ErrBadHex     = fmt.Errorf("%w (hex does not match pattern)", ErrInvalidNumValue)
	ErrBadOctal   = fmt.Errorf("%w (octal does not match pattern)", ErrInvalidNumValue)
	ErrBadBinary  = fmt.Errorf("%w (binary does not match pattern)", ErrInvalidNumValue)

	ErrEmptyNumber         = fmt.Errorf("%w (number is empty)", ErrInvalidNumValue)
	ErrSepsOnlyInDecimals  = fmt.Errorf("%w (separators available only in numbers base 10)", ErrInvalidNumValue)
	ErrTooManySepsInNumber = fmt.Errorf("%w (too many decimal separators)", ErrInvalidNumValue)
)

func readNumber(r *reader) (*big.Float, error) {

	length := 0
	var bytes []byte
	var err error

	for {

		length++

		bytes, err = r.peekN(length)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return big.NewFloat(0), err
		}

		ch := rune(bytes[len(bytes)-1])
		if ch == ';' || ch == '/' || unicode.IsSpace(ch) {
			bytes = bytes[0 : len(bytes)-1]
			break
		}
	}

	strOriginal := string(bytes)
	r.discard(length - 1)

	str := strOriginal
	if len(str) == 0 {
		return big.NewFloat(0), ErrEmptyNumber
	}

	sign := 0
	if str[0] == '-' {
		sign = -1
	} else if str[0] == '+' {
		sign = 1
	}

	if sign != 0 {
		str = str[1:]
	}

	base := 10
	if len(str) > 2 {
		if strings.HasPrefix(str, "0b") {
			base = 2
			if !patternBinary.MatchString(strOriginal) {
				return big.NewFloat(0), ErrBadBinary
			}
		} else if strings.HasPrefix(str, "0o") {
			base = 8
			if !patternOctal.MatchString(strOriginal) {
				return big.NewFloat(0), ErrBadOctal
			}
		} else if strings.HasPrefix(str, "0x") {
			base = 16
			if !patternHex.MatchString(strOriginal) {
				return big.NewFloat(0), ErrBadHex
			}
		}
	}

	if base == 10 {
		if !patternDecimal.MatchString(strOriginal) {
			return big.NewFloat(0), ErrBadDecimal
		}
	} else {
		str = str[2:]
		if strings.ContainsRune(str, '.') {
			return big.NewFloat(0), ErrSepsOnlyInDecimals
		}
	}

	if sign < 0 {
		str = "-" + str
	}

	str = strings.ReplaceAll(str, "_", "")
	if base == 10 {
		f, _, err := big.ParseFloat(str, 10, 53, big.AwayFromZero)
		return f, err
	}

	val, err := strconv.ParseInt(str, base, 64)
	return new(big.Float).SetInt64(val), err
}
