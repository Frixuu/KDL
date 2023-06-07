package kdl

import "bufio"

type writer struct {
	writer *bufio.Writer
	depth  int
}

func writeSpace(w *writer) error {
	return w.writer.WriteByte(' ')
}
