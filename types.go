package kdl

import (
	"math/big"
	"strconv"
	"strings"
)

type TypeTag int

const (
	TypeBool TypeTag = iota + 1
	TypeNumber
	TypeString
	TypeRawString
	TypeNull
	TypeDocument
	TypeDefault
	TypeObjects
)

type KDLValue struct {
	Type     TypeTag
	RawValue interface{}
	//declaredType string
}

func (v KDLValue) MustBool() bool {
	b, ok := v.RawValue.(bool)
	if !ok {
		panic("value was not a boolean")
	}
	return b
}

func (v KDLValue) MustNumber() *big.Float {
	n, ok := v.RawValue.(*big.Float)
	if !ok {
		panic("value was not a number")
	}
	return n
}

func (v KDLValue) MustString() string {
	s, ok := v.RawValue.(string)
	if !ok {
		panic("value was not a string")
	}
	return s
}

func (kdlValue KDLValue) RecreateKDL() (string, error) {
	switch kdlValue.Type {
	case TypeBool:
		return strconv.FormatBool(kdlValue.MustBool()), nil
	case TypeNumber:
		f64, _ := kdlValue.MustNumber().Float64()
		return strconv.FormatFloat(f64, 'f', -1, 64), nil
	case TypeString, TypeRawString:
		return RecreateString(kdlValue.MustString()), nil
	case TypeDocument:
		document := kdlValue.RawValue.([]KDLValue)
		var s strings.Builder
		for i, v := range document {

			str, err := v.RecreateKDL()
			if err != nil {
				return "", err
			}

			s.WriteString(str)
			if i+1 != len(document) {
				s.WriteRune(' ')
			}
		}
		return s.String(), nil
	case TypeNull:
		return "null", nil
	case TypeDefault:
		return "", nil
	case TypeObjects:
		objects := kdlValue.RawValue.([]KDLObject)
		if len(objects) < 1 {
			return "", nil
		}

		var s strings.Builder
		for _, obj := range objects {
			objStr, err := RecreateKDLObj(obj)
			if err != nil {
				return "", err
			}
			s.WriteString(objStr + "; ")
		}
		return "{ " + s.String() + "}", nil
	default:
		return "", ErrInvalidTypeTag
	}
}

func RecreateString(s string) string {
	return strings.ReplaceAll(strconv.Quote(s), "/", "\\/")
}

func (kdlValue KDLValue) ToString() (string, error) {
	switch kdlValue.Type {
	case TypeString, TypeRawString:
		s := kdlValue.RawValue.(string)
		return s, nil
	default:
		return kdlValue.RecreateKDL()
	}
}

type KDLObject interface {
	GetKey() string
	GetValue() KDLValue
}

func RecreateKDLObj(kdlObj KDLObject) (string, error) {
	s, err := kdlObj.GetValue().RecreateKDL()
	if err != nil {
		return "", nil
	}
	if len(s) > 0 {
		s = " " + s
	}
	key := kdlObj.GetKey()
	if strings.Contains(key, " ") || strconv.Quote(key) != "\""+key+"\"" {
		key = strconv.Quote(key)
	}
	return key + s, nil
}

type KDLBool struct {
	key   string
	value KDLValue
}

func NewKDLBool(key string, value bool) KDLBool {
	return KDLBool{key: key, value: KDLValue{Type: TypeBool, RawValue: value}}
}

func (kdlNode KDLBool) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLBool) GetValue() KDLValue {
	return kdlNode.value
}

type KDLNumber struct {
	key   string
	value KDLValue
}

func NewKDLNumber(key string, value float64) KDLNumber {
	return KDLNumber{key: key, value: KDLValue{Type: TypeNumber, RawValue: big.NewFloat(value)}}
}

func (kdlNode KDLNumber) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLNumber) GetValue() KDLValue {
	return kdlNode.value
}

type KDLString struct {
	key   string
	value KDLValue
}

func NewKDLString(key string, value string) KDLString {
	value = strings.ReplaceAll(value, "\n", "\\n")
	s, _ := strconv.Unquote(`"` + value + `"`)
	return KDLString{key: key, value: KDLValue{Type: TypeString, RawValue: s}}
}

func (kdlNode KDLString) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLString) GetValue() KDLValue {
	return kdlNode.value
}

type KDLRawString struct {
	key   string
	value KDLValue
}

func NewKDLRawString(key string, value string) KDLRawString {
	return KDLRawString{key: key, value: KDLValue{Type: TypeRawString, RawValue: value}}
}

func (kdlNode KDLRawString) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLRawString) GetValue() KDLValue {
	return kdlNode.value
}

type KDLDocument struct {
	key   string
	value KDLValue
}

func NewKDLDocument(key string, value []KDLValue) KDLDocument {
	return KDLDocument{key: key, value: KDLValue{Type: TypeDocument, RawValue: value}}
}

func (kdlNode KDLDocument) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLDocument) GetValue() KDLValue {
	return kdlNode.value
}

type KDLNull struct {
	key   string
	value KDLValue
}

func NewKDLNull(key string) KDLNull {
	return KDLNull{key: key, value: KDLValue{Type: TypeNull}}
}

func (kdlNode KDLNull) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLNull) GetValue() KDLValue {
	return kdlNode.value
}

type KDLDefault struct {
	key   string
	value KDLValue
}

func NewKDLDefault(key string) KDLDefault {
	return KDLDefault{key: key, value: KDLValue{Type: TypeDefault}}
}

func (kdlNode KDLDefault) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLDefault) GetValue() KDLValue {
	return kdlNode.value
}

type KDLObjects struct {
	key   string
	value KDLValue
}

func NewKDLObjects(key string, objects []KDLObject) KDLObjects {
	return KDLObjects{key: key, value: KDLValue{Type: TypeObjects, RawValue: objects}}
}

func (kdlNode KDLObjects) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLObjects) GetValue() KDLValue {
	return kdlNode.value
}

func (kdlObjs KDLObjects) ToObjMap() KDLObjectsMap {
	ret := make(KDLObjectsMap)
	objects := kdlObjs.value.RawValue.([]KDLObject)
	for _, obj := range objects {
		ret[obj.GetKey()] = obj
	}
	return ret
}

func (kdlObjs KDLObjects) ToValueMap() KDLValuesMap {
	ret := make(KDLValuesMap)
	objects := kdlObjs.value.RawValue.([]KDLObject)
	for _, obj := range objects {
		ret[obj.GetKey()] = obj.GetValue()
	}
	return ret
}

type KDLObjectsMap map[string]KDLObject
type KDLValuesMap map[string]KDLValue
