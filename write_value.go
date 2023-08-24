package kdl

import (
	"math"
	"math/big"
	"strconv"
	"strings"
)

var unescapeReplacer = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	"\n", `\n`,
	"\r", `\r`,
	"\t", `\t`,
	"\b", `\b`,
	"\f", `\f`,
)

func writeString(w *writer, s string) error {

	if err := w.writer.WriteByte('"'); err != nil {
		return err
	}
	if _, err := w.writer.WriteString(unescapeReplacer.Replace(s)); err != nil {
		return err
	}
	return w.writer.WriteByte('"')
}

func writeBool(w *writer, b bool) error {
	v := bytesFalse[:]
	if b {
		v = bytesTrue[:]
	}
	_, err := w.writer.Write(v)
	return err
}

func writeInteger(w *writer, i *big.Int) error {
	text := i.Text(10)

	if strings.HasSuffix(text, "000000") {
		exp := 6
		for {
			if text[len(text)-exp-1] != '0' {
				break
			}
			exp++
		}
		text = text[:len(text)-exp] + "E+" + strconv.Itoa(exp)
	}

	_, err := w.writer.WriteString(text)
	return err
}

func writeFloatNoExponent(w *writer, f *big.Float) (err error) {
	text := f.Text('f', 14)

	hadZeroes := false
	for {
		zero := strings.HasSuffix(text, "00")
		if !zero {
			break
		}
		hadZeroes = true
		text = text[:len(text)-1]
	}

	if hadZeroes && !strings.HasSuffix(text, ".0") {
		text = text[:len(text)-1]
	}

	_, err = w.writer.WriteString(text)
	return
}

var bigFloatZero = big.NewFloat(0.0)

func writeFloat(w *writer, f *big.Float) error {

	if f.Cmp(bigFloatZero) == 0 {
		_, err := w.writer.WriteString("0.0")
		return err
	}

	if f.IsInf() {
		_, err := w.writer.WriteString("Inf")
		return err
	}

	// Mode 'G' switches to sci mode later than we would like
	// and has troubles with choosing the right precision,
	// so we decide on form on our own

	d, _ := f.Float64()
	abs := math.Abs(d)
	if abs > 0.1_000_000 && abs < 1_000_000_000 {
		return writeFloatNoExponent(w, f)
	}

	text := f.Text('E', 15)
	man, exp, ok := strings.Cut(text, "E")
	if !ok {
		return writeFloatNoExponent(w, f)
	}

	manf, err := strconv.ParseFloat(man, 64)
	if err != nil {
		return err
	}

	err = writeFloatNoExponent(w, big.NewFloat(manf))
	if err != nil {
		return err
	}

	err = w.writer.WriteByte('E')
	if err != nil {
		return err
	}

	_, err = w.writer.WriteString(exp)
	return err
}

func writeNull(w *writer) error {
	_, err := w.writer.Write(bytesNull[:])
	return err
}

func writeValue(w *writer, v *Value) error {

	err := writeTypeHint(w, v.TypeHint)
	if err != nil {
		return err
	}

	switch v.Type {
	case TypeString:
		return writeString(w, v.StringValue())
	case TypeInteger:
		return writeInteger(w, v.IntegerValue())
	case TypeFloat:
		return writeFloat(w, v.FloatValue())
	case TypeBool:
		return writeBool(w, v.BoolValue())
	case TypeNull:
		return writeNull(w)
	default:
		return errInvalidTypeTag
	}
}

func writeIdentifier(w *writer, i Identifier) (err error) {
	if isAllowedBareIdentifier(string(i)) {
		_, err = w.writer.WriteString(string(i))
	} else {
		err = writeString(w, string(i))
	}
	return
}

// writeTypeHint writes a type hint to the output, if the hint is present.
func writeTypeHint(w *writer, hint TypeHint) error {

	if hint.IsAbsent() {
		return nil
	}

	if err := w.writer.WriteByte('('); err != nil {
		return err
	}

	if err := writeIdentifier(w, hint.MustGet()); err != nil {
		return err
	}

	return w.writer.WriteByte(')')
}
