package yamaha

import "bytes"

const whitespaces = "\t\n\v\f\r "

func isSpace(c byte) bool {
	return bytes.IndexByte([]byte(whitespaces), c) >= 0
}
func startsSpace(s []byte) bool {
	if len(s) == 0 {
		return false
	}
	return isSpace(s[0])
}
func trimSpace(s []byte) []byte {
	return bytes.TrimLeft(s, whitespaces)
}
func nextSpace(s []byte) int {
	return bytes.IndexAny(s, whitespaces)
}
func cutSpace(s []byte) ([]byte, []byte) {
	s = trimSpace(s)
	i := nextSpace(s)
	if i < 0 {
		return s, nil
	}
	return s[:i], s[i:]
}
func cutWord(s []byte) ([]byte, []byte) {
	// Note: Missing support for \" escaping, unlike documented by Yamaha
	// This is OK, as a value containing " was never observed in practice.
	if len(s) == 0 {
		return nil, nil
	}
	s = trimSpace(s)
	if s[0] != '"' {
		return cutSpace(s)
	} else {
		s = s[1:]
		i := bytes.IndexByte(s[:], '"')
		if i < 0 {
			return s, nil
		}
		return s[0:i], s[i+1:]
	}
}
