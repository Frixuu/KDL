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

var quotedReplacer = strings.NewReplacer(`\/`, `/`, "\n", `\n`)

func readQuotedString(r *reader) (string, error) {

	s, err := readQuotedStringInner(r)
	if err != nil {
		return s, err
	}

	return strconv.Unquote(`"` + quotedReplacer.Replace(s) + `"`)
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

var ErrExpectedString = fmt.Errorf("%w: expected string", ErrInvalidSyntax)

func readString(r *reader) (string, error) {
	ch, err := r.peek()
	if err != nil {
		return "", err
	}

	switch ch {
	case '"':
		return readQuotedString(r)
	case 'r':
		r.discard(1)
		return readRawString(r)
	default:
		return "", ErrExpectedString
	}
}

var bytesTrue = [...]byte{'t', 'r', 'u', 'e'}
var bytesFalse = [...]byte{'f', 'a', 'l', 's', 'e'}
var ErrExpectedBool = fmt.Errorf("%w: expected boolean", ErrInvalidSyntax)

func readBool(r *reader) (bool, error) {

	var expected []byte

	start, err := r.peek()
	if err != nil {
		return false, err
	}

	switch start {
	case 't':
		expected = bytesTrue[:]
	case 'f':
		expected = bytesFalse[:]
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

var bytesNull = [...]byte{'n', 'u', 'l', 'l'}
var ErrExpectedNull = fmt.Errorf("%w: expected null", ErrInvalidSyntax)

func readNull(r *reader) error {

	next, err := r.isNext(bytesNull[:])
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

var ErrInvalidBareIdentifier = fmt.Errorf("%w: not a valid bare identifier", ErrInvalidSyntax)

func readBareIdentifier(r *reader, stopOnCloseParen bool) (Identifier, error) {

	ch, err := r.peek()
	if err != nil {
		return "", err
	}

	if !isAllowedInitialCharacter(ch) {
		return "", ErrInvalidBareIdentifier
	}

	chars := make([]rune, 0, 16)
	chars = append(chars, ch)
	r.discard(1)

	for {

		ch, err := r.peek()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}

		if isWhitespace(ch) {
			break
		}

		if !isRuneAllowedInBareIdentifier(rune(ch)) {
			if stopOnCloseParen && ch == ')' {
				break
			}
			return "", ErrInvalidBareIdentifier
		}

		_, _ = r.readRune()
		chars = append(chars, ch)
	}

	ident := string(chars)
	// Sanity check: could be starting with -digit at this point
	if !isAllowedBareIdentifier(ident) {
		return "", ErrInvalidBareIdentifier
	}

	return Identifier(ident), nil
}

var ErrInvalidIdentifier = fmt.Errorf("%w: not a valid identifier", ErrInvalidSyntax)
var ErrInvalidInitialCharacter = fmt.Errorf("%w: not a valid initial character", ErrInvalidSyntax)

func readIdentifier(r *reader, stopOnCloseParen bool) (Identifier, error) {
	ch, err := r.peek()
	if err != nil {
		return "", err
	}

	if ch == '"' {
		s, err := readQuotedString(r)
		return Identifier(s), err
	}

	// r could mean a raw string or a bare ident, more checks are necessary
	if ch == 'r' {
		s, err := r.peekN(2)
		if err != nil {
			if errors.Is(err, io.EOF) {
				r.discard(1)
				return "r", nil
			}
			return "", nil
		}

		switch s[1] {
		case '"':
			r.discard(1)
			s, err := readRawString(r)
			return Identifier(s), err
		case '#':
			// TODO Still no idea, but assume it's a raw string for now
			r.discard(1)
			s, err := readRawString(r)
			return Identifier(s), err
		default:
			return readBareIdentifier(r, stopOnCloseParen)
		}
	}

	if isAllowedInitialCharacter(ch) {
		return readBareIdentifier(r, stopOnCloseParen)
	}

	return "", ErrInvalidInitialCharacter
}

func readTypeHint(r *reader) (Identifier, error) {

	ch, err := r.peek()
	if err != nil {
		return "", err
	}

	if ch != '(' {
		return "", nil
	}

	r.discard(1)

	id, err := readIdentifier(r, true)
	if err != nil {
		return "", err
	}

	ch, err = r.peek()
	if err != nil {
		return "", err
	}

	if ch == ')' {
		r.discard(1)
		return id, nil
	}

	return "", ErrInvalidSyntax
}
