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
	TypeNumber
	TypeString
)

type Value struct {
	Type     TypeTag
	TypeHint Identifier
	RawValue interface{}
}

func NewNullValue() Value {
	return Value{Type: TypeNull}
}

func NewBoolValue(v bool, hint Identifier) Value {
	return Value{Type: TypeBool, RawValue: v, TypeHint: hint}
}

func (v *Value) AsBool() bool {
	if v.Type != TypeBool {
		panic("value is not a boolean")
	}
	return v.RawValue.(bool)
}

func NewNumberValue(v *big.Float, hint Identifier) Value {
	return Value{Type: TypeNumber, RawValue: v, TypeHint: hint}
}

func (v *Value) AsNumber() *big.Float {
	if v.Type != TypeNumber {
		panic("value is not a number")
	}
	return v.RawValue.(*big.Float)
}

func NewStringValue(v string, hint Identifier) Value {
	return Value{Type: TypeString, RawValue: v, TypeHint: hint}
}

func (v *Value) AsString() string {
	if v.Type != TypeString {
		panic("value is not a string")
	}
	return v.RawValue.(string)
}
