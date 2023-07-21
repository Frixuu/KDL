package kdl

import (
	"bufio"
	"bytes"
	"io"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func readerFromString(s string) *reader {
	return wrapReader(bufio.NewReader(bytes.NewBufferString(s)))
}

func TestReadsQuotedString(t *testing.T) {

	reader := readerFromString(`"Hi!""Why, \"hello \nthere!""foo
	\n\"bar"extra`)

	s, err := readQuotedString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "Hi!", s)

	s, err = readQuotedString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "Why, \"hello \nthere!", s)

	s, err = readQuotedString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "foo\n\t\n\"bar", s)

	_, err = readQuotedString(reader)
	assert.ErrorIs(t, err, ErrInvalidSyntax)
}

func TestReadsRawString(t *testing.T) {

	reader := readerFromString(`r###"oh
	Hi"##there##!
"###r"extra data`)

	s, err := readRawString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "oh\n\tHi\"##there##!\n", s)

	_, err = readRawString(reader)
	assert.ErrorIs(t, err, io.EOF)

	reader = readerFromString(`r#"one pound"#`)
	s, err = readRawString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "one pound", s)

	reader = readerFromString(`r"no pounds"`)
	s, err = readRawString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "no pounds", s)

}

func TestReadsString(t *testing.T) {

	reader := readerFromString(`r##"foo"##"bar"`)

	s, err := readString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "foo", s)

	s, err = readString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "bar", s)
}

func TestReadsBool(t *testing.T) {

	reader := readerFromString("truefalsetent")

	b, err := readBool(reader)
	assert.NoError(t, err)
	assert.Equal(t, true, b)

	b, err = readBool(reader)
	assert.NoError(t, err)
	assert.Equal(t, false, b)

	_, err = readBool(reader)
	assert.ErrorIs(t, err, ErrInvalidSyntax)
}

func TestReadsNull(t *testing.T) {

	reader := readerFromString("null, or not")
	var err error

	err = readNull(reader)
	assert.NoError(t, err)

	err = readNull(reader)
	assert.ErrorIs(t, err, ErrInvalidSyntax)
}

func expectNumber(t *testing.T, r *reader, v float64) {
	n, err := readNumber(r)
	assert.NoError(t, err)
	x, _ := n.Float64()
	assert.InDelta(t, v, x, 0.0001)
	r.discardBytes(1)
}

func TestReadsNumberDecimal(t *testing.T) {
	reader := readerFromString("4 +2 -6 1_33_7 4e3 7e-2 -1.1e-2")
	expectNumber(t, reader, 4.0)
	expectNumber(t, reader, 2.0)
	expectNumber(t, reader, -6.0)
	expectNumber(t, reader, 1337.0)
	expectNumber(t, reader, 4000.0)
	expectNumber(t, reader, 0.07)
	expectNumber(t, reader, -0.011)
}

func TestReadsNumberHex(t *testing.T) {
	reader := readerFromString("0xc 0xa_0_f -0xD2")
	expectNumber(t, reader, 12.0)
	expectNumber(t, reader, 2575.0)
	expectNumber(t, reader, -210.0)
}

func TestReadsNumberOctal(t *testing.T) {
	reader := readerFromString("0o1_0 -0o26")
	expectNumber(t, reader, 8.0)
	expectNumber(t, reader, -22.0)
}

func TestReadsNumberBinary(t *testing.T) {
	reader := readerFromString("0b1 -0b1000_0001")
	expectNumber(t, reader, 1.0)
	expectNumber(t, reader, -129.0)
}

func TestReadsBareIdentifier(t *testing.T) {

	var errInvalidInitial *errInvalidInitialCharInBareIdent
	var errInvalidIdent *errInvalidBareIdent

	reader := readerFromString("abc")
	id, err := readBareIdentifier(reader, stopModeFreestanding)
	assert.NoError(t, err)
	assert.EqualValues(t, "abc", id)

	reader = readerFromString("def ")
	id, err = readBareIdentifier(reader, stopModeFreestanding)
	assert.NoError(t, err)
	assert.EqualValues(t, "def", id)

	reader = readerFromString("012")
	_, err = readBareIdentifier(reader, stopModeFreestanding)
	assert.ErrorAs(t, err, &errInvalidInitial)

	reader = readerFromString("-cool")
	id, err = readBareIdentifier(reader, stopModeFreestanding)
	assert.NoError(t, err)
	assert.EqualValues(t, "-cool", id)

	reader = readerFromString("-12")
	_, err = readBareIdentifier(reader, stopModeFreestanding)
	assert.ErrorAs(t, err, &errInvalidIdent)

	reader = readerFromString(`" `)
	_, err = readBareIdentifier(reader, stopModeFreestanding)
	assert.ErrorAs(t, err, &errInvalidInitial)
}

func TestReadsIdentifier(t *testing.T) {

	reader := readerFromString(`foo "bar baz" radio r#"gaga"#`)

	ident, err, _ := readIdentifier(reader, stopModeFreestanding)
	assert.NoError(t, err)
	assert.EqualValues(t, "foo", ident)

	_ = readUntilSignificant(reader)
	ident, err, _ = readIdentifier(reader, stopModeFreestanding)
	assert.NoError(t, err)
	assert.EqualValues(t, "bar baz", ident)

	_ = readUntilSignificant(reader)
	ident, err, _ = readIdentifier(reader, stopModeFreestanding)
	assert.NoError(t, err)
	assert.EqualValues(t, "radio", ident)

	_ = readUntilSignificant(reader)
	ident, err, _ = readIdentifier(reader, stopModeFreestanding)
	assert.NoError(t, err)
	assert.EqualValues(t, "gaga", ident)
}

func TestReadsTypeHint(t *testing.T) {

	reader := readerFromString("(foo)")
	hint, err := readMaybeTypeHint(reader)
	assert.NoError(t, err)
	assert.EqualValues(t, "foo", hint)

	reader = readerFromString("(bar baz)")
	_, err = readMaybeTypeHint(reader)
	assert.ErrorIs(t, err, ErrInvalidSyntax)

	reader = readerFromString("(\"hello world\")")
	hint, err = readMaybeTypeHint(reader)
	assert.NoError(t, err)
	assert.EqualValues(t, "hello world", hint)

	reader = readerFromString(`("hello\")`)
	_, err = readMaybeTypeHint(reader)
	assert.ErrorIs(t, err, io.EOF)

	reader = readerFromString("(aaaaa")
	_, err = readMaybeTypeHint(reader)
	assert.ErrorIs(t, err, io.EOF)
}

func TestReadsValue(t *testing.T) {

	reader := readerFromString(`true (temp)-3.5 ("hey")null "foo" what`)

	value, err := readValue(reader)
	assert.NoError(t, err)
	assert.EqualValues(t, NewBoolValue(true, ""), value)

	_ = readUntilSignificant(reader)
	value, err = readValue(reader)
	assert.NoError(t, err)
	// different rounding mode
	assert.EqualExportedValues(t, NewNumberValue(big.NewFloat(-3.5), "temp"), value)

	_ = readUntilSignificant(reader)
	value, err = readValue(reader)
	assert.NoError(t, err)
	assert.EqualValues(t, NewNullValue("hey"), value)

	_ = readUntilSignificant(reader)
	value, err = readValue(reader)
	assert.NoError(t, err)
	assert.EqualValues(t, NewStringValue("foo", ""), value)

	_ = readUntilSignificant(reader)
	_, err = readValue(reader)
	assert.Error(t, err)
}
