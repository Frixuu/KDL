package kdl

import (
	"math/big"
	"strconv"
	"strings"
)

func (value Value) RecreateKDL() (string, error) {
	switch value.Type {
	case TypeBool:
		return strconv.FormatBool(value.AsBool()), nil
	case TypeNumber:
		f64, _ := value.AsNumber().Float64()
		return strconv.FormatFloat(f64, 'f', -1, 64), nil
	case TypeString:
		return RecreateString(value.AsString()), nil
	case TypeDocument:
		document := value.RawValue.([]Value)
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
		objects := value.RawValue.([]KDLObject)
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

func (value Value) ToString() (string, error) {
	switch value.Type {
	case TypeString:
		s := value.RawValue.(string)
		return s, nil
	default:
		return value.RecreateKDL()
	}
}

type KDLObject interface {
	GetKey() string
	GetValue() Value
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
	value Value
}

func NewKDLBool(key string, value bool) KDLBool {
	return KDLBool{key: key, value: Value{Type: TypeBool, RawValue: value}}
}

func (kdlNode KDLBool) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLBool) GetValue() Value {
	return kdlNode.value
}

type KDLNumber struct {
	key   string
	value Value
}

func NewKDLNumber(key string, value float64) KDLNumber {
	return KDLNumber{key: key, value: Value{Type: TypeNumber, RawValue: big.NewFloat(value)}}
}

func (kdlNode KDLNumber) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLNumber) GetValue() Value {
	return kdlNode.value
}

type KDLString struct {
	key   string
	value Value
}

func NewKDLString(key string, value string) KDLString {
	value = strings.ReplaceAll(value, "\n", "\\n")
	s, _ := strconv.Unquote(`"` + value + `"`)
	return KDLString{key: key, value: Value{Type: TypeString, RawValue: s}}
}

func (kdlNode KDLString) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLString) GetValue() Value {
	return kdlNode.value
}

type KDLRawString struct {
	key   string
	value Value
}

func NewKDLRawString(key string, value string) KDLRawString {
	return KDLRawString{key: key, value: Value{Type: TypeString, RawValue: value}}
}

func (kdlNode KDLRawString) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLRawString) GetValue() Value {
	return kdlNode.value
}

type KDLDocument struct {
	key   string
	value Value
}

func NewKDLDocument(key string, value []Value) KDLDocument {
	return KDLDocument{key: key, value: Value{Type: TypeDocument, RawValue: value}}
}

func (kdlNode KDLDocument) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLDocument) GetValue() Value {
	return kdlNode.value
}

type KDLNull struct {
	key   string
	value Value
}

func NewKDLNull(key string) KDLNull {
	return KDLNull{key: key, value: Value{Type: TypeNull}}
}

func (kdlNode KDLNull) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLNull) GetValue() Value {
	return kdlNode.value
}

type KDLDefault struct {
	key   string
	value Value
}

func NewKDLDefault(key string) KDLDefault {
	return KDLDefault{key: key, value: Value{Type: TypeDefault}}
}

func (kdlNode KDLDefault) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLDefault) GetValue() Value {
	return kdlNode.value
}

type KDLObjects struct {
	key   string
	value Value
}

func NewKDLObjects(key string, objects []KDLObject) KDLObjects {
	return KDLObjects{key: key, value: Value{Type: TypeObjects, RawValue: objects}}
}

func (kdlNode KDLObjects) GetKey() string {
	return kdlNode.key
}

func (kdlNode KDLObjects) GetValue() Value {
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

func (kdlObjs KDLObjects) ToValueMap() ValuesMap {
	ret := make(ValuesMap)
	objects := kdlObjs.value.RawValue.([]KDLObject)
	for _, obj := range objects {
		ret[obj.GetKey()] = obj.GetValue()
	}
	return ret
}

type KDLObjectsMap map[string]KDLObject
type ValuesMap map[string]Value
