package kdl

import (
	"math/big"
)

type TypeTag int

const (
	TypeNull TypeTag = iota + 1
	TypeBool
	TypeNumber
	TypeString

	TypeDocument
	TypeDefault
	TypeObjects
)

type Value struct {
	Type     TypeTag
	TypeHint Identifier
	RawValue interface{}
}

func (v *Value) AsBool() bool {
	if v.Type != TypeBool {
		panic("value is not a boolean")
	}
	return v.RawValue.(bool)
}

func (v *Value) AsNumber() *big.Float {
	if v.Type != TypeNumber {
		panic("value is not a number")
	}
	return v.RawValue.(*big.Float)
}

func (v *Value) AsString() string {
	if v.Type != TypeString {
		panic("value is not a string")
	}
	return v.RawValue.(string)
}
