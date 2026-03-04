package parsers

import (
	"bufio"
	"bytes"
	"io"
)

func readAllAndRestore(r io.Reader) ([]byte, io.Reader, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return nil, nil, err
	}
	return b, bytes.NewReader(b), nil
}

func scanLines(r io.Reader, fn func(string)) error {
	s := bufio.NewScanner(r)
	for s.Scan() {
		fn(s.Text())
	}
	return s.Err()
}
