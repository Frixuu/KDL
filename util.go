package kdl

type pair[K comparable, V any] struct {
	key   K
	value V
}

// toPairs converts a `map[K]V` to a slice of `pair[K, V]` structs.
func toPairs[K comparable, V any](m map[K]V) []pair[K, V] {
	var pairs []pair[K, V] = make([]pair[K, V], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, pair[K, V]{key: k, value: v})
	}
	return pairs
}

func remainingUTF8Bytes(b byte) int {
	if b <= 0b0111_1111 {
		return 0
	} else if b >= 0b1100_0000 && b <= 0b1101_1111 {
		return 1
	} else if b <= 0b1110_1111 {
		return 2
	} else if b <= 0b1111_0111 {
		return 3
	} else {
		return -1
	}
}
