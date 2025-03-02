package yamaha

import "bytes"

const whitespaces = "\t\n\v\f\r "

// isSpace decides whether c is considered whitespace in Yamaha SCP.
//
// testing suggest this matches isspace() from stdc
func isSpace(c byte) bool {
	return bytes.IndexByte([]byte(whitespaces), c) >= 0
}

// startsSpace decides whether the first byte of s is whitespace.
func startsSpace(s []byte) bool {
	if len(s) == 0 {
		return false
	}
	return isSpace(s[0])
}

// trimSpace trims whitespace from the beginning of s.
func trimSpace(s []byte) []byte {
	return bytes.TrimLeft(s, whitespaces)
}

// nextSpace returns the index of the first whitespace in s.
//
// nextSpace returns -1 if there is no whitespace.
func nextSpace(s []byte) int {
	return bytes.IndexAny(s, whitespaces)
}

// cutSpace splits s into two parts upon the next whitespace.
//
// Any fronting whitespace is trimmed. If there is no whitespace, the whole
// slice is returned as the first part.
func cutSpace(s []byte) ([]byte, []byte) {
	s = trimSpace(s)
	i := nextSpace(s)
	if i < 0 {
		return s, nil
	}
	return s[:i], s[i:]
}

// cutWord splits s into two parts at the end of the first word.
//
// A word is a sequence of characters between " marks, otherwise a sequence of
// non-whitespace characters until the first whitespace or the end of string.
// Any fronting whitespace before the first word is trimmed.
func cutWord(s []byte) ([]byte, []byte) {
	// There is some word about Yamaha supporting \"-like escaping, but it did
	// not work when I tested them. Ignoring escaping here for preformance.
	// I strongly suggest NOT using " or \ in channel names or other values.
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
