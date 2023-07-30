package kdl

// remainingUTF8Bytes returns the count of the remaining continuation bytes,
// given a start of a multi-byte sequence.
// If b is US-ASCII, returns 0. If b is not valid, returns -1.
func remainingUTF8Bytes(b byte) int {
	if b <= 0b0111_1111 {
		return 0
	} else if b >= 0b1100_0000 {
		if b <= 0b1101_1111 {
			return 1
		} else if b <= 0b1110_1111 {
			return 2
		} else if b <= 0b1111_0111 {
			return 3
		}
	}
	return -1
}
