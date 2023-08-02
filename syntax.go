package kdl

import (
	"regexp"
	"unicode"
	"unicode/utf8"

	"golang.org/x/exp/slices"
)

// keywords are reserved symbols that cannot be used as bare identifiers.
var keywords = [...]string{"true", "false", "null"}

func isKeyword(s string) bool {
	return slices.Contains(keywords[:], s)
}

// charsSlashDash represents a sequence of bytes
// that tells the parser to discard the immediately following
// node, argument or property.
var charsSlashDash = [...]byte{'/', '-'}

var charsStartComment = [...]byte{'/', '/'}
var charsStartCommentBlock = [...]byte{'/', '*'}
var charsEndCommentBlock = [...]byte{'*', '/'}
var charsCRLF = [...]byte{'\r', '\n'}

// isNewLine checks if the rune is a line break character.
//
// Note: according to spec, CRLF is treated as a *singular* new line.
// This function does not check for it.
func isNewLine(ch rune) bool {
	if ch < 16 {
		return ch == '\n' || ch == '\r' || ch == 0xc
	}
	return ch == 0x85 || ch == 0x2028 || ch == 0x2029
}

var charsWhitespaceBig = [...]rune{
	0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200a,
	0x202f, 0x205f,
	0x3000,
}

// isWhitespace checks if the rune is a whitespace character.
func isWhitespace(ch rune) bool {
	if ch < 0x80 {
		return ch == 0x20 || ch == 0x9
	}
	if ch < 0x2000 {
		return ch == 0xa0 || ch == 0x1680
	}
	if ch > 0x3000 {
		return false
	}
	_, found := slices.BinarySearch(charsWhitespaceBig[:], ch)
	return found
}

// Identifier is a fancy name for a string
// in place of a node's name, type hint or a property key.
type Identifier string

func startsWithDigit(s string) bool {
	if len(s) < 1 {
		return false
	}

	r, size := utf8.DecodeRuneInString(s)
	if unicode.IsDigit(r) {
		return true
	}

	if r == '-' || r == '+' {
		if len(s) <= size {
			return false
		}
		n := s[size]
		if n >= '0' && n <= '9' {
			return true
		}

		if n >= 128 {
			r, _ = utf8.DecodeRuneInString(s[size:])
			return unicode.IsDigit(r)
		}

		return false
	}

	return false
}

var patternBareIdentifier = regexp.MustCompile(`^([^\/(){}<>;[\]=,"0-9\-+\s\\]|[\-+][^\/(){}<>;[\]=,"0-9\s\\])[^\/(){}<>;[\]=,"\s\\]*$`)

func isAllowedBareIdentifier(s string) bool {
	return !isKeyword(s) && patternBareIdentifier.MatchString(s)
}

var asciiAllowedInBareIdent = [128]byte{
	// 1  2  3  4  5  6  7  8  9  A  B  C  D  E  F
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 0x00 - 0x0F
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, // 0x10 - 0x1F
	1, 1, 0, 1, 1, 1, 1, 1, 0, 0, 1, 1, 0, 1, 1, 0, // 0x20 - 0x2F
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 0, 1, // 0x30 - 0x3F
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0x40 - 0x4F
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 0, 0, 1, 1, // 0x50 - 0x5F
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, // 0x60 - 0x6F
	1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 0, 1, 0, 1, 1, // 0x70 - 0x7F
}

func isRuneAllowedInBareIdentifier(ch rune) bool {
	if ch < 0x80 {
		return asciiAllowedInBareIdent[byte(ch)] > 0
	}
	return ch <= 0x10ffff
}

// isAllowedInitialCharacter checks if a bare identifier is allowed to start with this rune.
func isAllowedInitialCharacter(ch rune) bool {
	return isRuneAllowedInBareIdentifier(ch) && !unicode.IsDigit(ch)
}

func isValidValueTerminator(ch rune) bool {
	return ch == ';' || ch == '}' || isWhitespace(ch) || isNewLine(ch)
}
