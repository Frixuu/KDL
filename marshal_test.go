package kdl

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStructIntoChildren(t *testing.T) {
	s := struct {
		Foo string
		Bar string `kdl:"baz"`
	}{}

	c := marshalContext{}
	doc := NewDocument()
	structToChildren(&c, reflect.ValueOf(s), &doc)

	assert.Len(t, doc.Nodes, 2)
	assert.EqualValues(t, "foo", doc.Nodes[0].Name)
	assert.EqualValues(t, "baz", doc.Nodes[1].Name)
}

func TestStructIntoNode(t *testing.T) {
	s := struct {
		Foo  int `kdl:"-"`
		Bar  int
		Baz  float32 `kdl:",argument"`
		Quox string  `kdl:"hello,property"`
	}{}

	c := marshalContext{}
	n := NewNode("")
	structIntoNode(&c, reflect.ValueOf(s), &n)

	if assert.Len(t, n.Args, 1) {
		assert.EqualValues(t, TypeFloat, n.Args[0].Type)
	}

	if assert.Len(t, n.Props, 2) {
		assert.Equal(t, TypeInteger, n.GetProp("bar").Type)
		assert.Equal(t, TypeString, n.GetProp("hello").Type)
	}
}

func TestValueConverts(t *testing.T) {

	n := 3
	v, err := valueToKDLValue(reflect.ValueOf(n))
	assert.NoError(t, err)
	assert.EqualValues(t, n, v.AsInteger().Int64())

	s := "foo"
	v, err = valueToKDLValue(reflect.ValueOf(s))
	assert.NoError(t, err)
	assert.EqualValues(t, s, v.AsString())
}
