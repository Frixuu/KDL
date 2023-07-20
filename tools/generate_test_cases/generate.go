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

package kdl

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
			w.WriteString("\tdoc, err := ParseString(input)\n")
			w.WriteString("\tassert.NoError(t, err)\n")
			output, err := os.ReadFile(outputPath)
			panicOnError(err)
			w.WriteString("\toutput := `")
			w.Write(output)
			w.WriteString("`\n")
			w.WriteString("\twritten, err := doc.WriteString()\n")
			w.WriteString("\tassert.NoError(t, err)\n")
			w.WriteString("\tassert.Equal(t, output, written)\n")
		}

		w.WriteString("}\n")
	}
}
