package kdlgo

import (
	"path/filepath"
	"strconv"
	"testing"
)

//go:generate go run tools/generate_test_cases/generate.go

func TestParseFromFile(t *testing.T) {
	objs, err := ParseFile(filepath.Join("testdata", "test.kdl"))
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		`firstkey "first\n\ttab\nnewline\"\nval" "testing\""`,
		`numbers 543 234 85720394`,
		`thirdkey true null`,
		`secondkey 12 "test" null false "testagain"`,
		`anotherkey "true" 123.543 null true`,
		`moreKeys false true`,
		`keyonly`,
		`testcomment`,
		`objects { node1 12; node2 "string"; node3 null; }`,
		`multiline-node "random"`,
		`title "Some title"`,
		`"quoted node" "quoted value"`,
		`"quoted node for numbers" 21 43 465 "string"`,
		`smile "😁"`,
		`!@#$@$%Q#$%~@!40 "1.2.3" { !!!!! true; }`,
		`foo123~!@#$%^&*.:'|/?+ "weeee"`,
		`ノード { お名前 "☜(ﾟヮﾟ☜)"; }`,
		`foo { bar true; } "baz" { quux false; } 1 2 3`,
		`key "value"`,
		`test "value"`,
	}

	objects, _ := objs.GetValue().RawValue.([]KDLObject)
	if len(objects) != len(expected) {
		t.Fatal(
			"There should be " + strconv.Itoa(len(expected)) +
				" KDLObjects. Got " + strconv.Itoa(len(objects)) + " instead.",
		)
	}

	for i, obj := range objects {
		s, err := RecreateKDLObj(obj)
		if err != nil {
			t.Fatal(err)
			return
		}
		if s != expected[i] {
			t.Error(
				"Item number "+strconv.Itoa(i+1)+" is incorrectly parsed.\n",
				"Expected: '"+expected[i]+"' but got '"+s+"' instead",
			)
		}
	}
}
