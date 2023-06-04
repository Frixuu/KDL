package kdl

import (
	"strconv"
	"strings"
)

func checkQuotedString(s strings.Builder) string {
	ss := s.String()
	unquoted, err := strconv.Unquote(ss)
	if err != nil {
		return ss
	} else {
		return unquoted
	}
}
