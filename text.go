package kdl

import (
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// var caserTitle cases.Caser = cases.Title(language.English)
var caserLower cases.Caser = cases.Lower(language.English)
