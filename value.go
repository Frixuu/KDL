package kdl

import (
	"errors"
	"math/big"
)

// TypeTag discriminates between Value types.
type TypeTag byte

const (
	TypeInvalid TypeTag = iota // The described Value is in an invalid state.

	TypeNull    // The described Value holds a null.
	TypeBool    // The described Value holds a boolean.
	TypeString  // The described Value holds a string.
	TypeInteger // The described Value holds an integer.
	TypeFloat   // The described Value holds a floating point number.
)

var errInvalidTypeTag = errors.New("value has invalid type tag")

// Value can be used either as an argument or a property to a Node.
type Value struct {
	RawValue interface{}
	TypeHint TypeHint
	Type     TypeTag
}

// NewNullValue constructs a Value that holds a null.
func NewNullValue(hint TypeHint) Value {
	return Value{Type: TypeNull, TypeHint: hint}
}

// NewBoolValue constructs a Value that holds a boolean.
func NewBoolValue(v bool, hint TypeHint) Value {
	return Value{Type: TypeBool, RawValue: v, TypeHint: hint}
}

// AsBool returns the inner bool value or panics, if the Value is not a boolean.
func (v Value) AsBool() bool {
	if v.Type != TypeBool {
		panic("value is not a boolean")
	}
	return v.RawValue.(bool)
}

// NewStringValue constructs a Value that holds a string.
func NewStringValue(v string, hint TypeHint) Value {
	return Value{Type: TypeString, RawValue: v, TypeHint: hint}
}

// AsString returns the inner string value or panics, if the Value is not a string.
func (v Value) AsString() string {
	if v.Type != TypeString {
		panic("value is not a string")
	}
	return v.RawValue.(string)
}

// NewIntegerValue constructs a Value that holds an integer.
func NewIntegerValue(v *big.Int, hint TypeHint) Value {
	return Value{Type: TypeInteger, RawValue: v, TypeHint: hint}
}

// AsInteger returns the inner int value or panics, if the Value is not an integer.
func (v Value) AsInteger() *big.Int {
	if v.Type != TypeInteger {
		panic("value is not an integer")
	}
	return v.RawValue.(*big.Int)
}

// NewFloatValue constructs a Value that holds a float.
func NewFloatValue(v *big.Float, hint TypeHint) Value {
	return Value{Type: TypeFloat, RawValue: v, TypeHint: hint}
}

// AsFloat returns the inner float value or panics, if the Value is not a floating point number.
func (v Value) AsFloat() *big.Float {
	if v.Type != TypeFloat {
		panic("value is not a real number")
	}
	return v.RawValue.(*big.Float)
}

// newInvalidValue constructs a new Value that is in an invalid state.
func newInvalidValue() Value {
	return Value{Type: TypeInvalid}
}
