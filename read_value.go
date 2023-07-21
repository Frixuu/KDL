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

		bytes, err := r.peekBytes(count)
		if err != nil {
			r.discardBytes(count)
			return string(bytes), err
		}

		ch := rune(bytes[len(bytes)-1])
		if ch == '\\' {

			bs, err := r.peekBytes(count + 1)
			if err != nil {
				r.discardBytes(count)
				return string(bytes), err
			}

			next := bs[len(bs)-1] == byte('"')
			if next {
				count += 2
				continue
			}

		} else if ch == '"' {

			toRet := string(bytes[:len(bytes)-1])
			r.discardBytes(count)
			return toRet, nil
		}

		count++
	}
}

var errExpectedRawString = fmt.Errorf("%w: expected raw string", ErrInvalidSyntax)

func readRawString(r *reader) (string, error) {

	ch, err := r.peekRune()
	if err != nil {
		return "", err
	}

	// A raw string must start with an 'r'
	if ch != 'r' {
		return "", errExpectedRawString
	}

	// followed by 0 or more '#' characters
	leadingPoundCount := 0
	length := 2

	for {

		bytes, err := r.peekBytes(length)
		if err != nil {
			return "", err
		}

		ch := bytes[len(bytes)-1]
		if ch == '#' {
			leadingPoundCount++
			length++
		} else if ch == '"' {
			// and a doublequote.
			break
		} else {
			return "", errExpectedRawString
		}
	}

	// The string proper starts now
	contentStart := length
	closingPoundCount := 0
	isJustAfterDoublequotes := false
	var bytes []byte

	for {

		if isJustAfterDoublequotes && leadingPoundCount == closingPoundCount {
			r.discardBytes(length)
			return string(bytes[contentStart : len(bytes)-leadingPoundCount-1]), nil
		}

		length++
		bytes, err = r.peekBytes(length)
		if err != nil {
			return "", err
		}

		ch := bytes[len(bytes)-1]
		if ch == '"' {
			// The contents of the string may have possibly ended.
			// To return, we must now read the exact number of '#' characters
			// that we started the raw string with
			isJustAfterDoublequotes = true
			continue
		}

		if isJustAfterDoublequotes && ch == '#' {
			closingPoundCount++
		} else {
			closingPoundCount = 0
			isJustAfterDoublequotes = false
		}
	}
}

var ErrExpectedString = fmt.Errorf("%w: expected string", ErrInvalidSyntax)

func readString(r *reader) (string, error) {
	ch, err := r.peekRune()
	if err != nil {
		return "", err
	}

	switch ch {
	case '"':
		return readQuotedString(r)
	case 'r':
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

	start, err := r.peekRune()
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

		bytes, err = r.peekBytes(length)
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
	r.discardBytes(length - 1)

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

	ch, err := r.peekRune()
	if err != nil {
		return "", err
	}

	if !isAllowedInitialCharacter(ch) {
		return "", ErrInvalidBareIdentifier
	}

	chars := make([]rune, 0, 16)
	chars = append(chars, ch)
	r.discardBytes(1)

	for {

		ch, err := r.peekRune()
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
	ch, err := r.peekRune()
	if err != nil {
		return "", err
	}

	if ch == '"' {
		s, err := readQuotedString(r)
		return Identifier(s), err
	}

	// r could mean a raw string or a bare ident, more checks are necessary
	if ch == 'r' {
		s, err := readRawString(r)
		if err != nil {
			return readBareIdentifier(r, stopOnCloseParen)
		}
		return Identifier(s), err
	}

	if isAllowedInitialCharacter(ch) {
		return readBareIdentifier(r, stopOnCloseParen)
	}

	return "", ErrInvalidInitialCharacter
}

func readMaybeTypeHint(r *reader) (Identifier, error) {

	ch, err := r.peekRune()
	if err != nil {
		return "", err
	}

	if ch != '(' {
		return "", nil
	}

	r.discardBytes(1)

	id, err := readIdentifier(r, true)
	if err != nil {
		return "", err
	}

	ch, err = r.peekRune()
	if err != nil {
		return "", err
	}

	if ch == ')' {
		r.discardBytes(1)
		return id, nil
	}

	return "", ErrInvalidSyntax
}

func readValue(r *reader) (Value, error) {

	hint, err := readMaybeTypeHint(r)
	if err != nil {
		return newInvalidValue(), err
	}

	ch, err := r.peekRune()
	if err != nil {
		return newInvalidValue(), err
	}

	if unicode.IsDigit(ch) {
		v, err := readNumber(r)
		if err != nil {
			return newInvalidValue(), err
		}
		return NewNumberValue(v, hint), nil
	}

	switch ch {
	case '"':
		v, err := readQuotedString(r)
		if err != nil {
			return newInvalidValue(), err
		}
		return NewStringValue(v, hint), nil
	case 't', 'f':
		v, err := readBool(r)
		if err != nil {
			return newInvalidValue(), err
		}
		return NewBoolValue(v, hint), nil
	case '-', '+':
		v, err := readNumber(r)
		if err != nil {
			return newInvalidValue(), err
		}
		return NewNumberValue(v, hint), nil
	case 'r':
		v, err := readRawString(r)
		if err != nil {
			return newInvalidValue(), err
		}
		return NewStringValue(v, hint), nil
	case 'n':
		err := readNull(r)
		return NewNullValue(hint), err
	default:
		return newInvalidValue(), ErrInvalidSyntax
	}
}
