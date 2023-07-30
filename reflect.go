package kdl

import (
	"reflect"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// var caserTitle cases.Caser = cases.Title(language.English)
var caserLower cases.Caser = cases.Lower(language.English)

func determineName(sf reflect.StructField) string {
	name := caserLower.String(sf.Name)
	tag, ok := sf.Tag.Lookup("kdl")
	if ok {
		renamed, _, _ := strings.Cut(tag, ",")
		if renamed != "" {
			name = renamed
		}
	}
	return name
}
