package kdl

import (
	"regexp"

	"golang.org/x/exp/slices"
)

type Identifier string

var bareIdPattern = regexp.MustCompile(`^([^\/(){}<>;[\]=,"0-9\-+\s]|[\-+][^\/(){}<>;[\]=,"0-9\s])[^\/(){}<>;[\]=,"\s]*$`)
var bareIdForbiddenKeywords = []string{"true", "false"}

func isAllowedBareIdentifier(s string) bool {
	return !slices.Contains(bareIdForbiddenKeywords, s) &&
		bareIdPattern.MatchString(s)
}
