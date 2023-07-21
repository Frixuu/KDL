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

func writeNumber(w *writer, f *big.Float) error {
	_, err := w.writer.WriteString(f.Text('g', -1))
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
		return writeString(w, v.AsString())
	case TypeNumber:
		return writeNumber(w, v.AsNumber())
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
