package panasonic

import (
	"testing"
)

func Test_int2hex(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		length   int
		expected string
	}{
		{"Zero", 0, 2, "00"},
		{"Positive", 15, 2, "0F"},
		{"Negative", -5, 2, "00"},
		{"Overflow", 256, 2, "FF"},
		{"Length 1", 10, 1, "A"},
		{"Length 4", 4096, 4, "1000"},
		{"Zero Length", 15, 0, ""},
		{"Large Number", 65535, 4, "FFFF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := int2hex(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("Int2Hex(%d, %d) = %s; want %s", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}
func Test_int2dec(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		length   int
		expected string
	}{
		{"Single Digit", 5, 1, "5"},
		{"Two Digits", 42, 2, "42"},
		{"Padding Zeros", 7, 3, "007"},
		{"Max Value", 999, 3, "999"},
		{"Overflow", 1000, 3, "999"},
		{"Negative Input", -10, 2, "00"},
		{"Zero Input", 0, 4, "0000"},
		{"Large Number", 12345, 6, "012345"},
		{"Zero Length", 123, 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := int2dec(tt.input, tt.length)
			if result != tt.expected {
				t.Errorf("int2dec(%d, %d) = %s; want %s", tt.input, tt.length, result, tt.expected)
			}
		})
	}
}
func Test_hex2int(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Empty String", "", 0},
		{"Single Digit", "5", 5},
		{"Lowercase Hex", "a", 0},
		{"Uppercase Hex", "F", 15},
		{"Mixed Case", "Ab", 0},
		{"Multiple Digits", "123", 291},
		{"Max Valid Length", "FFFFFFF", 268435455},
		{"Exceeds Max Length", "FFFFFFFF", 0},
		{"Invalid Character", "G", 0},
		{"Mixed Valid and Invalid", "12G4", 0},
		{"All Zeros", "0000", 0},
		{"Leading Zeros", "00FF", 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hex2int(tt.input)
			if result != tt.expected {
				t.Errorf("hex2int(%s) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}
func Test_dec2int(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected int
	}{
		{"Single Digit", "7", 7},
		{"Multiple Digits", "123", 123},
		{"Leading Zeros", "0042", 42},
		{"Max Valid Length", "999999999", 999999999},
		{"Exceeds Max Length", "1000000000", 0},
		{"Empty String", "", 0},
		{"Non-Digit Character", "12a34", 0},
		{"Negative Sign", "-123", 0},
		{"Decimal Point", "12.34", 0},
		{"Whitespace", " 123 ", 0},
		{"All Zeros", "000000000", 0},
		{"Mixed Valid and Invalid", "123abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := dec2int(tt.input)
			if result != tt.expected {
				t.Errorf("dec2int(%s) = %d; want %d", tt.input, result, tt.expected)
			}
		})
	}
}
func Test_match(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		s       string
		want    bool
	}{
		{"Empty strings", "", "", true},
		{"Exact match", "abc", "abc", true},
		{"Pattern shorter", "ab", "abc", false},
		{"String shorter", "abc", "ab", false},
		{"Wildcard match", "\x00\x00\x00", "aB0", true},
		{"Partial wildcard", "a\x00c", "abc", true},
		{"Wildcard mismatch", "a\x00c", "a-d", false},
		{"Invalid set", "\x03", "a", false},
		{"Mixed wildcards and exact", "a\x00c\x01e", "abcDe", true},
		{"Mixed wildcards and exact mismatch", "a\x00c\x01e", "abcde", false},
		{"All wildcards", "\x00\x01\x02\x01", "aB2D", true},
		{"Complex pattern", "a\x00C\x01\x02F", "aXCD3F", true},
		{"Complex pattern mismatch", "a\x00C\x01\x02f", "aBCD3g", false},
		{"Magic stop char match", "A\x00C\x01E\x7F", "ABCDEFGHIJKL", true},
		{"Magic stop char mismatch", "A\x00C\x01E\x7F", "ABCDeFGHIJKL", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := match(tt.pattern, tt.s); got != tt.want {
				t.Errorf("match(%q, %q) = %v, want %v", tt.pattern, tt.s, got, tt.want)
			}
		})
	}
}
