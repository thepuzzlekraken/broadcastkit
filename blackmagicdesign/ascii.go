package blackmagicdesign

import (
	"bytes"
	"cmp"
	"iter"
	"strconv"
	"strings"
)

// whitespaces lists all characters considered whitespace by BMD protocols
// this conveniently coincides with the C isspace() function
const whitespaces = "\t\n\v\f\r "

// blankSplitter splits on blank lines containing at most whitespace
func blankSplitter(data []byte, eof bool) (advance int, token []byte, err error) {
	if eof && len(data) == 0 {
		return 0, nil, nil
	}

	for advance < len(data) {
		i := bytes.IndexByte(data[advance:], '\n')
		if i == -1 {
			break
		}
		advance += i + 1
		for j := advance; j < len(data); j++ {
			if data[j] == '\n' {
				return j + 1, data[0:advance], nil
			}
			if strings.IndexByte(whitespaces, data[j]) == -1 {
				advance = j + 1
				break
			}
		}
	}

	if eof {
		// Last command is incomplete, drop it
		return len(data), nil, nil
	}

	// Request more data.
	return 0, nil, nil
}

// colonLines splits ASCII lines on ':' into key-value pairs
func colonLines(b []byte) iter.Seq2[[]byte, []byte] {
	return func(yield func(key, value []byte) bool) {
		for _, l := range bytes.Split(b, []byte("\n")) {
			key, value, ok := bytes.Cut(l, []byte(":"))
			if !ok {
				continue
			}
			key = trimLeft(key)
			value = trim(value)
			if !yield(key, value) {
				return
			}
		}
	}
}

// numberedLines splits ASCII lines on ' ' into row-number value pairs
func numberedLines(b []byte) iter.Seq2[int, []byte] {
	return func(yield func(key int, value []byte) bool) {
		for _, l := range bytes.Split(b, []byte("\n")) {
			key, value, ok := bytes.Cut(l, []byte(" "))
			if !ok {
				continue
			}
			no, err := strconv.Atoi(string(trim(key)))
			if err != nil {
				continue
			}
			value = trim(value)
			if !yield(no, value) {
				return
			}
		}
	}
}

// orderedIter iterates over a map in key ascending order
func orderedIter[K cmp.Ordered, V any](m map[K]V) iter.Seq2[K, V] {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return func(yield func(key K, value V) bool) {
		for _, k := range keys {
			if !yield(k, m[k]) {
				return
			}
		}
	}
}

// trim whitespaces around s
func trim(b []byte) []byte {
	return bytes.Trim(b, whitespaces)
}

// trim whitespaces around s
func trimLeft(b []byte) []byte {
	return bytes.TrimLeft(b, whitespaces)
}

// uppercase does an in-place ASCII-only uppercase conversion of b
func uppercase(b []byte) {
	for i := 0; i < len(b); i++ {
		c := b[i]
		if 'a' <= c && c <= 'z' {
			c -= 'a' - 'A'
		}
		b[i] = c
	}
}

// lowercase does an in-place ASCII-only lowercase conversion of b
func lowercase(b []byte) {
	for i := 0; i < len(b); i++ {
		c := b[i]
		if 'A' <= c && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
}
