package kdl

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math/big"
	"reflect"
	"strconv"
	"strings"

	"golang.org/x/exp/slices"
)

type marshalContext struct {
	chain []reflect.Value
}

var errCannotMarshalType = errors.New("cannot marshal type (only structs and maps are supported)")

type errMarshalCycleDetected struct {
	chain []reflect.Value
	d     reflect.Value
}

func (e *errMarshalCycleDetected) Error() string {

	d := e.d
	chain := e.chain
	if chain == nil || !d.IsValid() {
		return "(cannot get error message as ErrCycleDetected is in invalid state)"
	}

	n := strings.ToUpper(d.Type().Name())
	var b strings.Builder
	b.WriteString("cycle detected when marshalling KDL: ")
	for i, v := range chain {
		b.WriteString(strconv.Itoa(i))
		b.WriteString(") ")
		if d == v {
			b.WriteString(n)
		} else {
			b.WriteString(v.Type().Name())
		}
		b.WriteString(" -> ")
	}
	b.WriteByte('[')
	b.WriteString(n)
	b.WriteByte(']')
	return b.String()
}

func marshal(v any) ([]byte, error) {
	var buf bytes.Buffer
	var data []byte
	err := marshalWriter(v, &buf)
	if err != nil {
		data = buf.Bytes()
	}
	return data, err
}

func marshalWriter(v any, w io.Writer) error {

	doc := NewDocument()

	chain := make([]reflect.Value, 0, 8)
	c := marshalContext{chain}

	if err := valueToChildren(&c, reflect.ValueOf(v), &doc); err != nil {
		return err
	}

	return doc.Write(w)
}

func tryPushChain(c *marshalContext, v reflect.Value) error {
	if slices.Contains(c.chain, v) {
		return &errMarshalCycleDetected{chain: c.chain, d: v}
	}
	c.chain = append(c.chain, v)
	return nil
}

func popChain(c *marshalContext) error {
	chain := c.chain
	c.chain = chain[:len(chain)-1]
	return nil
}

func valueToChildren(c *marshalContext, v reflect.Value, p nodeParent) (err error) {

	if err = tryPushChain(c, v); err != nil {
		return
	}

	switch v.Kind() {
	case reflect.Struct:
		err = structToChildren(c, v, p)
	case reflect.Map:
		err = mapToChildren(c, v, p)
	case reflect.Pointer:
		elem := v.Elem()
		if elem.Kind() == reflect.Struct {
			err = structToChildren(c, elem, p)
		} else {
			err = errCannotMarshalType
		}
	default:
		err = errCannotMarshalType
	}

	if err == nil {
		err = popChain(c)
	}

	return
}

func structToChildren(c *marshalContext, s reflect.Value, p nodeParent) error {

	t := s.Type()
	structFields := reflect.VisibleFields(t)
	for _, sf := range structFields {

		v := s.FieldByIndex(sf.Index)
		name := caserLower.String(sf.Name)

		tag, ok := sf.Tag.Lookup("kdl")
		if ok {
			opts := strings.Split(tag, ",")
			if opts[0] == "-" {
				continue
			}
			name = opts[0]
		}

		n := NewNode(name)
		if err := valueIntoNode(c, v, &n); err != nil {
			return err
		}

		p.AddChild(n)
	}

	return nil
}

type purpose int

const (
	purposeArgument purpose = iota
	purposeProperty
	purposeChildren
)

func structIntoNode(c *marshalContext, s reflect.Value, n *Node) error {

	if err := tryPushChain(c, s); err != nil {
		return err
	}

	t := s.Type()
	structFields := reflect.VisibleFields(t)

	childrenTaken := false

	for _, sf := range structFields {

		v := s.FieldByIndex(sf.Index)
		determinedName := caserLower.String(sf.Name)
		purpose := purposeProperty
		tag, ok := sf.Tag.Lookup("kdl")
		if ok {
			opts := strings.Split(tag, ",")

			if opts[0] == "-" {
				continue
			}

			determinedName = opts[0]
			if len(opts) > 1 {

				if slices.Contains(opts[1:], "argument") {
					purpose = purposeArgument
				} else if slices.Contains(opts[1:], "children") {
					if childrenTaken {
						return errors.New("this struct already defined one of its fields as children")
					}
					purpose = purposeChildren
					childrenTaken = true
				}
			}
		}

		switch purpose {
		case purposeArgument:
			val, err := valueToKDLValue(v)
			if err != nil {
				return err
			}
			n.AddArg(val)
		case purposeProperty:
			val, err := valueToKDLValue(v)
			if err != nil {
				return err
			}
			n.SetProp(Identifier(determinedName), val)
		case purposeChildren:
			if err := valueToChildren(c, v, n); err != nil {
				return err
			}
		}
	}

	return popChain(c)
}

func valueToKDLValue(v reflect.Value) (Value, error) {

	switch v.Kind() {
	case reflect.String:
		return NewStringValue(v.Interface().(string), NoHint()), nil
	case reflect.Bool:
		return NewBoolValue(v.Interface().(bool), NoHint()), nil
	case reflect.Float32, reflect.Float64:
		return NewFloatValue(big.NewFloat(v.Float()), NoHint()), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return NewIntegerValue(big.NewInt(v.Int()), NoHint()), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		b := new(big.Int)
		return NewIntegerValue(b.SetUint64(v.Uint()), NoHint()), nil
	case reflect.Pointer:
		if v.Pointer() == 0 {
			return NewNullValue(NoHint()), nil
		}
		return valueToKDLValue(v.Elem())
	}
	return newInvalidValue(), errors.New("invalid value kind")
}

var errBadMapKey = errors.New("only maps with string or stringer keys are supported")

func mapToChildren(c *marshalContext, m reflect.Value, p nodeParent) error {

	iter := m.MapRange()
	for iter.Next() {

		k := iter.Key()
		var name string
		if k.Kind() == reflect.String {
			name = k.Interface().(string)
		} else {
			s, ok := k.Interface().(fmt.Stringer)
			if !ok {
				return errBadMapKey
			}
			name = s.String()
		}

		v := iter.Value()

		n := NewNode(name)
		if err := valueIntoNode(c, v, &n); err != nil {
			return err
		}

		p.AddChild(n)
	}

	return nil
}

func valueIntoNode(c *marshalContext, v reflect.Value, n *Node) error {

	if err := tryPushChain(c, v); err != nil {
		return err
	}

	return popChain(c)
}
