package kdl

import (
	"bufio"
	"bytes"
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

	reader := readerFromString(`###"oh
	Hi"##there##!
"### extra data`)

	s, err := readRawString(reader)
	assert.NoError(t, err)
	assert.Equal(t, "oh\n\tHi\"##there##!\n", s)

	_, err = readRawString(reader)
	assert.ErrorIs(t, err, ErrInvalidSyntax)
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
	r.discard(1)
}

func TestReadsNumberDecimal(t *testing.T) {
	reader := readerFromString("4 +2 -6 1_33_7 4e3 7e-2 -1.1e2.2")
	expectNumber(t, reader, 4.0)
	expectNumber(t, reader, 2.0)
	expectNumber(t, reader, -6.0)
	expectNumber(t, reader, 1337.0)
	expectNumber(t, reader, 4000.0)
	expectNumber(t, reader, 0.07)
	//expectNumber(t, reader, -174.33825117)
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
