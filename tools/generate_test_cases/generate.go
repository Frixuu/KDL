package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {

	casesDir := filepath.Join("testdata", "kdl", "tests", "test_cases")
	inputDir := filepath.Join(casesDir, "input")
	outputDir := filepath.Join(casesDir, "expected_kdl")

	inputFiles, err := os.ReadDir(inputDir)
	panicOnError(err)

	f, err := os.Create("kdl_test.go")
	panicOnError(err)
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	w.WriteString(`// Code generated automatically. DO NOT EDIT.

package kdlgo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)
`)

	caser := cases.Title(language.English)
	for _, inputFile := range inputFiles {

		name := inputFile.Name()
		nameTrimmed := strings.TrimSuffix(name, ".kdl")

		if name == nameTrimmed {
			continue
		}

		var namePascalCase strings.Builder
		nameWords := strings.Split(nameTrimmed, "_")
		for _, word := range nameWords {
			namePascalCase.WriteString(caser.String(word))
		}

		w.WriteString("\nfunc Test")
		w.WriteString(namePascalCase.String())
		w.WriteString("(t *testing.T) {\n")
		w.WriteString("\tinput := `")

		input, err := os.ReadFile(filepath.Join(inputDir, name))
		panicOnError(err)
		w.Write(input)
		w.WriteString("`\n")

		outputPath := filepath.Join(outputDir, name)
		if _, err := os.Stat(outputPath); err != nil {
			w.WriteString("\t_, err := ParseString(input)\n")
			w.WriteString("\tassert.Error(t, err)\n")
		} else {
			w.WriteString("\t_, err := ParseString(input)\n")
			w.WriteString("\tassert.NoError(t, err)\n")
		}

		w.WriteString("}\n")
	}

	/*
			files, err := ioutil.ReadDir("testdata/kdl/tests/test_cases/input")
			if err != nil {
				log.Fatal(err)
			}

			for _, file := range files {
				name := file.Name()
				if !strings.HasSuffix(name, ".kdl") {
					continue
				}

				objs, _ := kdlgo.ParseFile("../tests/kdls/" + name)

				f, _ := os.Create("../tests/testers/" + strings.ReplaceAll(name, ".kdl", "_test.go"))
				defer f.Close()

				f.WriteString(`
		package kdlgo

		import (
			"strconv"
			"testing"
		)

		func Test` + strings.ToUpper(strings.Join(strings.Split(strings.TrimRight(name, ".kdl"), "_"), "")) + `(t *testing.T) {
		`)
				f.WriteString("	objs, err := kdlgo.ParseFile(\"../kdls/" + name + "\")")
				f.WriteString(`
			if err != nil {
				t.Fatal(err)
			}
			expected := []string{
		`)
				for _, obj := range objs.GetValue().Objects {
					s, _ := kdlgo.RecreateKDLObj(obj)
					f.WriteString("		`" + s + "`,")
				}

				f.WriteString(`
			}

			if len(objs.GetValue().Objects) != len(expected) {
				t.Fatal(
					"There should be " + strconv.Itoa(len(expected)) +
						" KDLObjects. Got " + strconv.Itoa(len(objs.GetValue().Objects)) + " instead.",
				)
			}

			for i, obj := range objs.GetValue().Objects {
				s, err := kdlgo.RecreateKDLObj(obj)
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
		`)
			}
	*/

}
