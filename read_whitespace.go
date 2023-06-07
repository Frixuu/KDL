package kdl

var whitespaceChars = [...]rune{
	0x9, 0x20, 0xa0, 0x1680,
	0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200a,
	0x202f, 0x205f,
	0x3000,
}

func isWhitespace(r *reader) (bool, error) {

	ch, err := r.readRune()
	if err != nil {
		return false, err
	}

	for _, ws := range whitespaceChars {
		if ch == ws {
			return true, nil
		}
	}

	return false, nil
}
