package panasonic

// FuseSet is a binary field of 128 bits
// FuseSet is little-endian: index 0 is the least significant part.
type FuseSet [4]uint32

func (f FuseSet) Set(i uint8) FuseSet {
	f[i/32] |= 1 << (i % 32)
	return f
}
func (f FuseSet) Clear(i uint8) FuseSet {
	f[i/32] &= ^(1 << (i % 32))
	return f
}
func (f FuseSet) Has(i uint8) bool {
	return (f[i/32] & (1 << (i % 32))) != 0
}
func (f FuseSet) Diff(g FuseSet) FuseSet {
	f[0] ^= g[0]
	f[1] ^= g[1]
	f[2] ^= g[2]
	f[3] ^= g[3]
	return f
}
func (f FuseSet) Union(g FuseSet) FuseSet {
	f[0] |= g[0]
	f[1] |= g[1]
	f[2] |= g[2]
	f[3] |= g[3]
	return f
}
func (f FuseSet) Intersection(g FuseSet) FuseSet {
	f[0] &= g[0]
	f[1] &= g[1]
	f[2] &= g[2]
	f[3] &= g[3]
	return f
}
func (f FuseSet) Invert() FuseSet {
	f[0] = ^f[0]
	f[1] = ^f[1]
	f[2] = ^f[2]
	f[3] = ^f[3]
	return f
}
func (f FuseSet) ShiftLeft(n uint) FuseSet {
	for n > 32 {
		f[3] = f[2]
		f[2] = f[1]
		f[1] = f[0]
		f[0] = 0
		n -= 32
	}
	f[3] <<= n
	f[3] |= (f[2] >> (32 - n))
	f[2] <<= n
	f[2] |= (f[1] >> (32 - n))
	f[1] <<= n
	f[1] |= (f[0] >> (32 - n))
	f[0] <<= n
	return f
}
func (f FuseSet) ShiftRight(n uint) FuseSet {
	for n > 32 {
		f[0] = f[1]
		f[1] = f[2]
		f[2] = f[3]
		f[3] = 0
		n -= 32
	}
	f[0] >>= n
	f[0] |= (f[1] << (32 - n))
	f[1] >>= n
	f[1] |= (f[2] << (32 - n))
	f[2] >>= n
	f[2] |= (f[3] << (32 - n))
	f[3] >>= n
	return f
}

// charSet is a clever bitmask to check for ASCII char existence in a set.
// Inspired by strings.asciiSet
type charSet [8]uint32

// makeCharSet creates a byteSet from a string of bytes.
func makeCharSet(chars string) charSet {
	var s charSet
	for i := range len(chars) {
		c := chars[i]
		s[c/32] |= 1 << (c % 32)
	}
	return s
}

var _ = makeCharSet // avoid "declared and not used" warnings

// contains reports whether c is inside the set.
func (s *charSet) contains(c byte) bool {
	return (s[c/32] & (1 << (c % 32))) != 0
}

// matchSets is a predefined list of charSets used by the Panasonic AW protocol
var matchSets = [...]charSet{
	{0x0, 0x07ff0000, 0x87fffffe, 0x07fffffe, 0x0, 0x0, 0x0, 0x0}, // makeCharSet("0123456789:ABCDEFGHIJKLMNOPQRSTUVWXYZ_abcdefghijklmnopqrstuvwxyz"),
	{0x0, 0x03ff0000, 0x0000007e, 0x00000000, 0x0, 0x0, 0x0, 0x0}, // makeCharSet("0123456789ABCDEF"),
	{0x0, 0x03ff0000, 0x00000000, 0x00000000, 0x0, 0x0, 0x0, 0x0}, // makeCharSet("0123456789"),
	{0x0, 0x00000000, 0x00000020, 0x00000020, 0x0, 0x0, 0x0, 0x0}, // makeCharSet("Ee"),
	{0x0, 0xfffffffe, 0xffffffff, 0x7fffffff, 0x0, 0x0, 0x0, 0x0}, // makeCharSet("!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~")
}

const (
	anySet = 0 // matchSets[anySet]: any Panasonic acceptable character
	hexSet = 1 // matchSets[hexSet]: hexadecimal value input characters
	decSet = 2 // matchSets[decSet]: decimal value input characters
	errSet = 3 // matchSets[errSet]: the letters e or E
	prtSet = 4 // matchSets[prtSet]: all printable ASCII characters
)

// hex2int converts fixed-length hex strings to integer.
// only accepts uppercase letters ABCDEF and digits 0123456789
// len(h) must be between 0 and 7
func hex2int(h string) int {
	if len(h) > 7 {
		return 0
	}
	var i int = 0
	for p := range len(h) {
		c := h[p]
		if !matchSets[hexSet].contains(c) {
			return 0
		}
		i <<= 4
		if c < 'A' {
			i |= int(c - '0')
		} else {
			i |= int(c - 'A' + 10)
		}
	}
	return i
}

// dec2int converts fixed-length decimal strings to integer.
// len(d) must be between 0 and 9
func dec2int(d string) int {
	if len(d) > 9 {
		return 0
	}
	var i int = 0
	for p := range len(d) {
		c := d[p]
		if !matchSets[decSet].contains(c) {
			return 0
		}
		i *= 10
		i += int(c - '0')
	}
	return i
}

// hexAlphabet are the chars used by the Panasonic AW protocol for values
// same as matchSets[hexSet]
const hexAlphabet = "0123456789ABCDEF"

// int2hex converts integer to fixed-length hex strings
// len(return) = l, prefixed with 0s
// l must be between 0 and 7
func int2hex(i int, l int) string {
	if l < 1 || l > 7 {
		return ""
	}
	if i < 0 {
		i = 0
	}
	m := 1 << (4 * l)
	if i >= m {
		i = m - 1
	}
	b := make([]byte, l, 7) // capacity defined to avoid escape to heap
	for p := len(b) - 1; p >= 0; p-- {
		b[p] = hexAlphabet[i&0xF]
		i = i >> 4
	}
	return string(b)
}

// int2dec converts integer to fixed-length decimal strings
// len(return) = l, prefixed with 0s
// l must be between 0 and 9
func int2dec(i int, l int) string {
	if l < 1 || l > 9 {
		return ""
	}
	if i < 0 {
		i = 0
	}
	m := 1
	for range l {
		m *= 10
	}
	if i >= m {
		i = m - 1
	}
	b := make([]byte, l, 9) // capacity defined to avoid escape to heap
	for p := len(b) - 1; p >= 0; p-- {
		b[p] = hexAlphabet[i%10]
		i /= 10
	}
	return string(b)
}

// match is the simplest possible pattern matcher for AWCommand parsing
// pattern is an ASCII string, which is matched against s as follows:
// - printable need to match exactly, case-sensitive
// - \x00 matches exactly one character from the acceptableSet [0-9A-Za-z:_]
// - \x01 matches exactly one character from the hexSet [0-9A-F]
// - \x02 matches exactly one character from the decSet [0-9]
// - \x03 matches exactly one characted from the errSet [eE]
// - \x04 matches exactly one printable ASCII character [ -~]
// - \x7F matches immediately - should be the last char of the pattern .+
// - any other character is invalid and behavior is undefined
func match(pattern string, s string) bool {
	if len(pattern) > len(s) {
		return false
	}
	for p := range len(pattern) {
		c := pattern[p]
		if c >= 32 {
			// exact match
			if c != s[p] {
				return c == '\xF7' // magic stop char check
			}
			continue
		}
		if c >= byte(len(matchSets)) {
			// invalid set
			return false
		}
		if !matchSets[c].contains(s[p]) {
			return false
		}
	}
	return len(pattern) == len(s)
}

// trim anything non-printable
func trim(b []byte) []byte {
	for len(b) > 0 {
		if matchSets[prtSet].contains(b[0]) {
			break
		}
		b = b[1:]
	}
	for len(b) > 0 {
		if matchSets[prtSet].contains(b[len(b)-1]) {
			break
		}
		b = b[:len(b)-1]
	}
	return b
}
