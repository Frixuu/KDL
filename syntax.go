package kdl

import (
	"regexp"
	"unicode"

	"golang.org/x/exp/slices"
)

// keywords are reserved symbols that cannot be used as bare identifiers.
var keywords = [...]string{"true", "false", "null"}

// charsSlashDash represents a sequence of bytes
// that tells the parser to discard the immediately following
// node, argument or property.
var charsSlashDash = [...]byte{'/', '-'}

var charsStartComment = [...]byte{'/', '/'}
var charsStartCommentBlock = [...]byte{'/', '*'}
var charsEndCommentBlock = [...]byte{'*', '/'}
var charsCRLF = [...]byte{'\r', '\n'}

// charsNewLine are all the runes that should be treated as new lines.
var charsNewLine = [...]rune{'\n', '\r', 0x85, 0xc, 0x2028, 0x2029}

// isNewLine checks if the rune is a line break character.
//
// Note: according to spec, CRLF is treated as a *singular* new line.
// This function does not check for it.
func isNewLine(ch rune) bool {
	return slices.Contains(charsNewLine[:], ch)
}

// charsWhitespace are all the runes that should be treated as whitespace.
var charsWhitespace = [...]rune{
	0x9, 0x20, 0xa0, 0x1680,
	0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200a,
	0x202f, 0x205f,
	0x3000,
}

// isWhitespace checks if the rune is a whitespace character.
func isWhitespace(ch rune) bool {
	return slices.Contains(charsWhitespace[:], ch)
}

type Identifier string

var patternBareIdentifier = regexp.MustCompile(`^([^\/(){}<>;[\]=,"0-9\-+\s]|[\-+][^\/(){}<>;[\]=,"0-9\s])[^\/(){}<>;[\]=,"\s]*$`)

func isAllowedBareIdentifier(s string) bool {
	return !slices.Contains(keywords[:], s) && patternBareIdentifier.MatchString(s)
}

var charsNotInBareIdentifier = [...]rune{
	'(', ')', '{', '}', '[', ']',
	'/', '\\', '<', '>', ';', '=', ',', '"',
}

func isRuneAllowedInBareIdentifier(ch rune) bool {
	return !slices.Contains(charsNotInBareIdentifier[:], ch) && ch > 0x20 && ch <= 0x10ffff
}

// isAllowedInitialCharacter checks if a bare identifier is allowed to start with this rune.
func isAllowedInitialCharacter(ch rune) bool {
	return !unicode.IsDigit(ch) && isRuneAllowedInBareIdentifier(ch)
}

func isValidValueTerminator(ch rune) bool {
	return ch == ';' || ch == '}' || isWhitespace(ch) || isNewLine(ch)
}
