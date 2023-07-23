package kdl

import (
	"errors"
	"math/big"
)

type TypeTag int

var ErrInvalidTypeTag = errors.New("value has invalid KDL type tag")

const (
	TypeInvalid TypeTag = iota

	TypeNull
	TypeBool
	TypeString
	TypeInteger
	TypeFloat
)

type Value struct {
	Type     TypeTag
	TypeHint TypeHint
	RawValue interface{}
}

func NewNullValue(hint TypeHint) Value {
	return Value{Type: TypeNull, TypeHint: hint}
}

func NewBoolValue(v bool, hint TypeHint) Value {
	return Value{Type: TypeBool, RawValue: v, TypeHint: hint}
}

func (v *Value) AsBool() bool {
	if v.Type != TypeBool {
		panic("value is not a boolean")
	}
	return v.RawValue.(bool)
}
func NewStringValue(v string, hint TypeHint) Value {
	return Value{Type: TypeString, RawValue: v, TypeHint: hint}
}

func (v *Value) AsString() string {
	if v.Type != TypeString {
		panic("value is not a string")
	}
	return v.RawValue.(string)
}

func NewIntegerValue(v *big.Int, hint TypeHint) Value {
	return Value{Type: TypeInteger, RawValue: v, TypeHint: hint}
}

func (v *Value) AsInteger() *big.Int {
	if v.Type != TypeInteger {
		panic("value is not an integer")
	}
	return v.RawValue.(*big.Int)
}

func NewFloatValue(v *big.Float, hint TypeHint) Value {
	return Value{Type: TypeFloat, RawValue: v, TypeHint: hint}
}

func (v *Value) AsFloat() *big.Float {
	if v.Type != TypeFloat {
		panic("value is not a real number")
	}
	return v.RawValue.(*big.Float)
}

func newInvalidValue() Value {
	return Value{Type: TypeInvalid}
}
