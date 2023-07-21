package kdl

import (
	"math/big"
	"strings"
)

var writeStringReplacer = strings.NewReplacer(
	"\n", `\n`,
	`\`, `\\`,
	`"`, `\"`,
)

func writeString(w *writer, s string) error {
	s = writeStringReplacer.Replace(s)
	_, err := w.writer.WriteString(`"` + s + `"`)
	return err
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
	_, err := w.writer.WriteString(i.Text(10))
	return err
}

func writeFloat(w *writer, f *big.Float) (err error) {
	text := f.Text('g', -1)
	_, err = w.writer.WriteString(text)
	if err == nil && !strings.ContainsAny(text, ".eE") {
		_, err = w.writer.WriteString(".0")
	}
	return
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
		return writeString(w, v.AsString())
	case TypeInteger:
		return writeInteger(w, v.AsInteger())
	case TypeFloat:
		return writeFloat(w, v.AsFloat())
	case TypeBool:
		return writeBool(w, v.AsBool())
	case TypeNull:
		return writeNull(w)
	default:
		return ErrInvalidTypeTag
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

func writeTypeHint(w *writer, hint Identifier) (err error) {

	if hint == "" {
		return nil
	}

	err = w.writer.WriteByte('(')
	if err != nil {
		return
	}

	err = writeIdentifier(w, hint)
	if err != nil {
		return
	}

	err = w.writer.WriteByte(')')
	return
}
