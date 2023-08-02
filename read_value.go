package kdl

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"

	"golang.org/x/exp/slices"
)

var unicodeEscapePattern = regexp.MustCompile(`\\u\{([0-9a-fA-F]{1,6})\}`)

func unicodeUnescapeFunc(matched string) string {
	i, _ := strconv.ParseInt(matched[3:len(matched)-1], 16, 32)
	return string(rune(i))
}

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

	str, escapes, err := readQuotedStringInner(r)
	if err != nil {
		return str, err
	}

	if escapes {
		str = unicodeEscapePattern.ReplaceAllStringFunc(str, unicodeUnescapeFunc)
		str = escapeReplacer.Replace(str)
	}

	return str, nil
}

var errUnexpectedEOFInsideString = fmt.Errorf("%w: did you forget to close a string?", ErrUnexpectedEOF)
var errExpectedQuotedString = fmt.Errorf("%w: expected quoted string", ErrInvalidSyntax)

func readQuotedStringInner(r *reader) (string, bool, error) {

	start, err := r.readByte()
	if err != nil {
		// EOF expected to be handled by the caller
		return "", false, err
	}

	if start != '"' {
		return "", false, errExpectedQuotedString
	}

	count := 1
	hasEscapes := false

	for {

		bytes, err := r.peekBytes(count)
		if err != nil {
			if err == io.EOF {
				err = errUnexpectedEOFInsideString
			}
			return string(bytes), hasEscapes, err
		}

		ch := bytes[len(bytes)-1]
		if ch == '\\' {

			hasEscapes = true

			bs, err := r.peekBytes(count + 1)
			if err != nil {
				if err == io.EOF {
					err = errUnexpectedEOFInsideString
				}
				return string(bytes), hasEscapes, err
			}

			escaped := bs[len(bs)-1]
			if escaped == '"' || escaped == '\\' {
				count += 2
				continue
			}

		} else if ch == '"' {

			toRet := string(bytes[:len(bytes)-1])
			r.discardBytes(count)
			return toRet, hasEscapes, nil
		}

		count++
	}
}

var errExpectedRawString = fmt.Errorf("%w: expected raw string", ErrInvalidSyntax)

func readRawString(r *reader) (string, error) {

	ch, err := r.peekByte()
	if err != nil {
		// EOF expected to be handled by the caller
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
			s := string(bytes[contentStart : len(bytes)-leadingPoundCount-1])
			r.discardBytes(length)
			return s, nil
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
			closingPoundCount = 0
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

var errExpectedString = fmt.Errorf("%w: expected string", ErrInvalidSyntax)

func readString(r *reader) (string, error) {

	ch, err := r.peekByte()
	if err != nil {
		return "", err
	}

	switch ch {
	case '"':
		return readQuotedString(r)
	case 'r':
		return readRawString(r)
	default:
		return "", errExpectedString
	}
}

var bytesTrue = [...]byte{'t', 'r', 'u', 'e'}
var bytesFalse = [...]byte{'f', 'a', 'l', 's', 'e'}
var errExpectedBool = fmt.Errorf("%w: expected boolean", ErrInvalidSyntax)

func readBool(r *reader) (bool, error) {

	var expected []byte

	start, err := r.peekByte()
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
		r.discardBytes(len(expected))
		return start == 't', nil
	}

	return false, errExpectedBool
}

var bytesNull = [...]byte{'n', 'u', 'l', 'l'}
var errExpectedNull = fmt.Errorf("%w: expected null", ErrInvalidSyntax)

func readNull(r *reader) error {

	next, err := r.isNext(bytesNull[:])
	if err != nil {
		return err
	}

	if next {
		r.discardBytes(4)
		return nil
	}

	return errExpectedNull
}

var (

	// Note: Patterns below do not support signs before the number: we're stripping them first

	patternDecimal = regexp.MustCompile(`^[0-9][_0-9]*(\.[0-9][_0-9]*)?([eE][-+]?[0-9][_0-9]*)?$`)
	errBadDecimal  = fmt.Errorf("%w (decimal does not match pattern)", errInvalidNumValue)

	patternHex = regexp.MustCompile(`^0x[0-9a-fA-F][_0-9a-fA-F]*$`)
	errBadHex  = fmt.Errorf("%w (hex does not match pattern)", errInvalidNumValue)
	prefixHex  = []byte{'0', 'x'}

	patternOctal = regexp.MustCompile(`^0o[0-7][_0-7]*$`)
	errBadOctal  = fmt.Errorf("%w (octal does not match pattern)", errInvalidNumValue)
	prefixOctal  = []byte{'0', 'o'}

	patternBinary = regexp.MustCompile(`^0b[01][_01]*$`)
	errBadBinary  = fmt.Errorf("%w (binary does not match pattern)", errInvalidNumValue)
	prefixBinary  = []byte{'0', 'b'}

	errInvalidNumValue    = fmt.Errorf("%w: bad numeric value", ErrInvalidSyntax)
	errEmptyNumber        = fmt.Errorf("%w (number is empty)", errInvalidNumValue)
	errSepsOnlyInDecimals = fmt.Errorf("%w (separators available only in numbers base 10)", errInvalidNumValue)

	errFailedToParseInt   = fmt.Errorf("%w (could not parse integer)", errInvalidNumValue)
	errFailedToParseFloat = fmt.Errorf("%w (could not parse float)", errInvalidNumValue)
)

type number struct {
	Value interface{}
	Type  TypeTag
}

func readNumber(r *reader) (number, error) {

	length := 0
	var data []byte
	var err error

	for {

		length++

		data, err = r.peekBytes(length)
		if err != nil {
			if err == io.EOF {
				break
			}
			return number{}, err
		}

		ch := rune(data[len(data)-1])
		if ch == ';' || ch == '/' || unicode.IsSpace(ch) {
			data = data[0 : len(data)-1]
			break
		}
	}

	if len(data) == 0 {
		return number{}, errEmptyNumber
	}

	sign := 0
	if data[0] == '-' {
		sign = -1
	} else if data[0] == '+' {
		sign = 1
	}

	if sign != 0 {
		data = data[1:]
	}

	base := 10
	if len(data) > 2 {
		maybeBasePrefix := data[0:2]
		if bytes.Equal(maybeBasePrefix, prefixBinary) {
			base = 2
			if !patternBinary.Match(data) {
				return number{}, errBadBinary
			}
		} else if bytes.Equal(maybeBasePrefix, prefixOctal) {
			base = 8
			if !patternOctal.Match(data) {
				return number{}, errBadOctal
			}
		} else if bytes.Equal(maybeBasePrefix, prefixHex) {
			base = 16
			if !patternHex.Match(data) {
				return number{}, errBadHex
			}
		}
	}

	if base == 10 {
		if !patternDecimal.Match(data) {
			return number{}, errBadDecimal
		}
	} else {
		data = data[2:]
		if slices.Contains(data, '.') {
			return number{}, errSepsOnlyInDecimals
		}
	}

	str := string(data)
	r.discardBytes(length - 1)

	str = strings.ReplaceAll(str, "_", "")
	if base == 10 {
		if strings.ContainsRune(str, '.') {
			f, _, err := big.ParseFloat(str, 10, 53, big.AwayFromZero)
			if err != nil {
				return number{}, errFailedToParseFloat
			}
			if sign < 0 {
				f = f.Neg(f)
			}
			return number{Type: TypeFloat, Value: f}, nil
		}
		if strings.ContainsAny(str, "eE") {
			str = strings.ToUpper(str)
			man, exp, _ := strings.Cut(str, "E")
			if strings.HasPrefix(exp, "-") {
				f, _, err := big.ParseFloat(str, 10, 53, big.AwayFromZero)
				if err != nil {
					return number{}, errFailedToParseFloat
				}
				if sign < 0 {
					f = f.Neg(f)
				}
				return number{Type: TypeFloat, Value: f}, nil
			} else {
				e, _ := strconv.Atoi(exp)
				str = man + strings.Repeat("0", e)
			}
		}
	}

	// Numbers in other bases are guaranteed to be integers
	i := new(big.Int)
	_, ok := i.SetString(str, base)
	if ok {
		if sign < 0 {
			i = i.Neg(i)
		}
		return number{Type: TypeInteger, Value: i}, nil
	}

	return number{}, errFailedToParseInt
}

var (
	errInvalidBareIdent              = fmt.Errorf("%w: invalid bare identifier", ErrInvalidSyntax)
	errInvalidCharInBareIdent        = fmt.Errorf("%w (illegal character)", errInvalidBareIdent)
	errInvalidInitialCharInBareIdent = fmt.Errorf("%w (does not start with a valid character)", errInvalidBareIdent)
)

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
		return "", errInvalidInitialCharInBareIdent
	}

	lengthBytes := 0
	for {

		b, err := r.peekBytes(lengthBytes + 1)
		if err != nil {
			if err == io.EOF {
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
				if err == io.EOF {
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
			return "", errInvalidCharInBareIdent
		}

		lengthBytes += (runeRemLen + 1)
	}

	b, err := r.peekBytes(lengthBytes)
	if err != nil {
		return "", err
	}

	// Unsafe string to avoid allocations if this was not a valid identifier
	ident := unsafe.String(unsafe.SliceData(b), len(b))
	if isKeyword(ident) {
		return "", errInvalidBareIdent
	}
	if startsWithDigit(ident) {
		return "", errInvalidBareIdent
	}

	// Actually make a copy now
	ident = string(b)
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
		err = errInvalidInitialCharInBareIdent
	}

	return
}

var errExpectedCloseHint = fmt.Errorf("%w: expected ) after type hint", ErrInvalidSyntax)

// readMaybeTypeHint reads an optional type hint, if one exists in the input.
func readMaybeTypeHint(r *reader) (TypeHint, error) {

	ch, err := r.peekByte()
	if err != nil {
		// EOF expected to be handled by the caller
		return NoHint(), err
	}

	if ch != '(' {
		// No hint in the input
		return NoHint(), nil
	}

	r.discardByte()

	// An identifier should follow right after - no whitespace nor comments
	ident, err, _ := readIdentifier(r, stopModeCloseParen)
	if err != nil {
		return NoHint(), err
	}

	// The parenthesis also should close just after - no whitespace nor comments
	ch, err = r.peekByte()
	if err != nil {
		if err == io.EOF {
			return NoHint(), ErrUnexpectedEOF
		}
		return NoHint(), err
	}

	if ch == ')' {
		r.discardByte()
		return Hint(string(ident)), nil
	}

	return NoHint(), errExpectedCloseHint
}

var errExpectedValue = fmt.Errorf("%w: expected value", ErrInvalidSyntax)

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
			return newInvalidValue(), errInvalidNumValue
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
			return newInvalidValue(), errInvalidNumValue
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
		return newInvalidValue(), errExpectedValue
	}
}
