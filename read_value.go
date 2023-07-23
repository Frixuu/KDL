package kdl

import (
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

var escapeReplacer = strings.NewReplacer(
	`\/`, `/`,
	`\\`, `\`,
	`\"`, `"`,
	`\n`, "\n",
	`\r`, "\r",
	`\t`, "\t",
	`\b`, "\b",
	`\f`, "\f",
)

func readQuotedString(r *reader) (string, error) {

	s, err := readQuotedStringInner(r)
	if err != nil {
		return s, err
	}

	return escapeReplacer.Replace(s), nil
}

var errExpectedQuotedString = fmt.Errorf("%w: expected quoted string", ErrInvalidSyntax)

func readQuotedStringInner(r *reader) (string, error) {

	count := 1

	start, err := r.readRune()
	if err != nil {
		return "", err
	}

	if start != '"' {
		return "", errExpectedQuotedString
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
var errExpectedBool = fmt.Errorf("%w: expected boolean", ErrInvalidSyntax)

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
		return false, errExpectedBool
	}

	next, err := r.isNext(expected)
	if err != nil {
		return false, err
	}

	if next {
		if start == 't' {
			r.discardBytes(4)
			return true, nil
		} else {
			r.discardBytes(5)
			return false, nil
		}
	}

	return false, errExpectedBool
}

var bytesNull = [...]byte{'n', 'u', 'l', 'l'}
var ErrExpectedNull = fmt.Errorf("%w: expected null", ErrInvalidSyntax)

func readNull(r *reader) error {

	next, err := r.isNext(bytesNull[:])
	if err != nil {
		return err
	}

	if next {
		r.discardBytes(4)
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

	errFailedToParseInt   = fmt.Errorf("%w (could not parse integer)", ErrInvalidNumValue)
	errFailedToParseFloat = fmt.Errorf("%w (could not parse float)", ErrInvalidNumValue)
)

type Number struct {
	Type  TypeTag
	Value interface{}
}

func readNumber(r *reader) (Number, error) {

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
			return Number{}, err
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
		return Number{}, ErrEmptyNumber
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
				return Number{}, ErrBadBinary
			}
		} else if strings.HasPrefix(str, "0o") {
			base = 8
			if !patternOctal.MatchString(strOriginal) {
				return Number{}, ErrBadOctal
			}
		} else if strings.HasPrefix(str, "0x") {
			base = 16
			if !patternHex.MatchString(strOriginal) {
				return Number{}, ErrBadHex
			}
		}
	}

	if base == 10 {
		if !patternDecimal.MatchString(strOriginal) {
			return Number{}, ErrBadDecimal
		}
	} else {
		str = str[2:]
		if strings.ContainsRune(str, '.') {
			return Number{}, ErrSepsOnlyInDecimals
		}
	}

	if sign < 0 {
		str = "-" + str
	}

	str = strings.ReplaceAll(str, "_", "")
	if base == 10 && strings.ContainsAny(str, ".eE") {
		f, _, err := big.ParseFloat(str, 10, 53, big.AwayFromZero)
		if err != nil {
			return Number{}, errFailedToParseFloat
		}
		return Number{Type: TypeFloat, Value: f}, nil
	}

	// Numbers in other bases are guaranteed to be integers
	i := new(big.Int)
	_, ok := i.SetString(str, base)
	if ok {
		return Number{Type: TypeInteger, Value: i}, nil
	}

	return Number{}, errFailedToParseInt
}

type errInvalidBareIdent struct {
	ident string
}

func (e *errInvalidBareIdent) Error() string {
	return fmt.Sprintf(
		"%s: \"%s\" is not a valid bare identifier",
		ErrInvalidSyntax.Error(),
		e.ident,
	)
}

func (e *errInvalidBareIdent) Unwrap() error {
	return ErrInvalidSyntax
}

type errInvalidCharInBareIdent struct {
	ch rune
}

func (e *errInvalidCharInBareIdent) Error() string {
	return fmt.Sprintf(
		"%s: '%s' is not a valid character in a bare identifier",
		ErrInvalidSyntax.Error(),
		string(e.ch),
	)
}

func (e *errInvalidCharInBareIdent) Unwrap() error {
	return ErrInvalidSyntax
}

type errInvalidInitialCharInBareIdent struct {
	ch rune
}

func (e *errInvalidInitialCharInBareIdent) Error() string {
	return fmt.Sprintf(
		"%s: '%s' is not a valid initial character for a bare identifier",
		ErrInvalidSyntax.Error(),
		string(e.ch),
	)
}

func (e *errInvalidInitialCharInBareIdent) Unwrap() error {
	return ErrInvalidSyntax
}

type identStopMode int

const (
	stopModeFreestanding identStopMode = iota
	stopModeCloseParen
	stopModeEquals
	stopModeSemicolon
)

func readBareIdentifier(r *reader, stopMode identStopMode) (Identifier, error) {

	ch, err := r.peekRune()
	if err != nil {
		return "", err
	}

	if !isAllowedInitialCharacter(ch) {
		return "", &errInvalidInitialCharInBareIdent{ch: ch}
	}

	lengthBytes := 0
	for {

		b, err := r.peekBytes(lengthBytes + 1)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return "", err
		}

		lastByte := b[len(b)-1]
		if !utf8.RuneStart(lastByte) {
			return "", ErrInvalidEncoding
		}

		runeRemLen := remainingUTF8Bytes(lastByte)
		if runeRemLen > 0 {
			b, err = r.peekBytes(lengthBytes + runeRemLen + 1)
			if err != nil {
				if errors.Is(err, io.EOF) {
					return "", ErrUnexpectedEOF
				}
				return "", err
			}
		}

		ch, _ := utf8.DecodeLastRune(b)
		if ch == utf8.RuneError {
			return "", ErrInvalidEncoding
		}

		if isWhitespace(ch) || isNewLine(ch) {
			break
		}

		if !isRuneAllowedInBareIdentifier(ch) {
			if stopMode == stopModeCloseParen && ch == ')' {
				break
			} else if stopMode == stopModeEquals && ch == '=' {
				break
			} else if stopMode == stopModeSemicolon && ch == ';' {
				break
			}
			return "", &errInvalidCharInBareIdent{ch: ch}
		}

		lengthBytes += (runeRemLen + 1)
	}

	b, err := r.peekBytes(lengthBytes)
	if err != nil {
		return "", err
	}

	ident := string(b)
	// Validate, could still be a keyword
	if !isAllowedBareIdentifier(ident) {
		return "", &errInvalidBareIdent{ident: ident}
	}

	r.discardBytes(lengthBytes)
	return Identifier(ident), nil
}

func readIdentifier(r *reader, stopMode identStopMode) (i Identifier, err error, quoted bool) {

	i = ""

	var ch rune
	ch, err = r.peekRune()
	if err != nil {
		return
	}

	var s string
	if ch == '"' {
		quoted = true
		s, err = readQuotedString(r)
		i = Identifier(s)
		return
	}

	// r could mean a raw string or a bare ident
	if ch == 'r' {
		s, err = readRawString(r)
		if err != nil {
			i, err = readBareIdentifier(r, stopMode)
			return
		}

		quoted = true
		i = Identifier(s)
		return
	}

	if isAllowedInitialCharacter(ch) {
		i, err = readBareIdentifier(r, stopMode)
	} else {
		err = &errInvalidInitialCharInBareIdent{ch: ch}
	}

	return
}

var errExpectedCloseHint = fmt.Errorf("%w: expected ) after type hint", ErrInvalidSyntax)

func readMaybeTypeHint(r *reader) (Identifier, error) {

	ch, err := r.peekRune()
	if err != nil {
		return "", err
	}

	if ch != '(' {
		return "", nil
	}

	r.discardBytes(1)

	id, err, _ := readIdentifier(r, stopModeCloseParen)
	if err != nil {
		return "", err
	}

	ch, err = r.peekRune()
	if err != nil {
		return "", err
	}

	if ch == ')' {
		r.discardBytes(1)
		return Identifier(id), nil
	}

	return "", errExpectedCloseHint
}

type errExpectedValue struct {
	found rune
}

func (e *errExpectedValue) Error() string {
	return fmt.Sprintf(
		"%s: expected value, got character '%s' which is not valid",
		ErrInvalidSyntax.Error(),
		string(e.found),
	)
}

func (e *errExpectedValue) Unwrap() error {
	return ErrInvalidSyntax
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
		n, err := readNumber(r)
		if err != nil {
			return newInvalidValue(), err
		}

		switch n.Type {
		case TypeFloat:
			return NewFloatValue(n.Value.(*big.Float), hint), nil
		case TypeInteger:
			return NewIntegerValue(n.Value.(*big.Int), hint), nil
		default:
			return newInvalidValue(), ErrInvalidNumValue
		}
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
		n, err := readNumber(r)
		if err != nil {
			return newInvalidValue(), err
		}

		switch n.Type {
		case TypeFloat:
			return NewFloatValue(n.Value.(*big.Float), hint), nil
		case TypeInteger:
			return NewIntegerValue(n.Value.(*big.Int), hint), nil
		default:
			return newInvalidValue(), ErrInvalidNumValue
		}
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
		return newInvalidValue(), &errExpectedValue{found: ch}
	}
}
