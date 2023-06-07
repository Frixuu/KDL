package kdl

var crlf = []byte{'\r', '\n'}
var newLineRunes = [...]rune{'\r', '\n', 0x85, 0xc, 0x2028, 0x2029}

func isNewLine(r *reader) (bool, error) {

	// CRLF is a special case where we match two codepoints,
	// so we have to deal with it first
	isCrlf, err := r.isNext(crlf)
	if err == nil && isCrlf {
		return true, nil
	}

	current, err := r.readRune()
	if err != nil {
		return false, err
	}

	for _, newLine := range newLineRunes {
		if current == newLine {
			return true, nil
		}
	}

	return false, nil
}
