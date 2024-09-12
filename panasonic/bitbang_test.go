package panasonic

import (
	"encoding/binary"
	"math/rand/v2"
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
func TestFuseSet_Set(t *testing.T) {
	tests := []struct {
		name     string
		initial  Bits128
		index    uint8
		expected Bits128
	}{
		{
			name:     "Set first bit",
			initial:  Bits128{0, 0},
			index:    0,
			expected: Bits128{1, 0},
		},
		{
			name:     "Set last bit in lo element",
			initial:  Bits128{0, 0},
			index:    63,
			expected: Bits128{1 << 63, 0},
		},
		{
			name:     "Set first bit in hi element",
			initial:  Bits128{0, 0},
			index:    64,
			expected: Bits128{0, 1},
		},
		{
			name:     "Set bit in middle of hi element",
			initial:  Bits128{0, 0},
			index:    85,
			expected: Bits128{0, 1 << 21},
		},
		{
			name:     "Set bit that's already set",
			initial:  Bits128{1 << 5, 0},
			index:    5,
			expected: Bits128{1 << 5, 0},
		},
		{
			name:     "Set with multiple bits",
			initial:  Bits128{1, 32768},
			index:    33,
			expected: Bits128{8589934593, 32768},
		},
		{
			name:     "Set last possible bit",
			initial:  Bits128{0, 0},
			index:    127,
			expected: Bits128{0, 1 << 63},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.Set(tt.index)
			if result != tt.expected {
				t.Errorf("FuseSet.Set(%d) = %v; want %v", tt.index, result, tt.expected)
			}
		})
	}
}
func TestFuseSet_Clear(t *testing.T) {
	tests := []struct {
		name     string
		initial  Bits128
		index    uint8
		expected Bits128
	}{
		{
			name:     "Clear first bit",
			initial:  Bits128{1, 0},
			index:    0,
			expected: Bits128{0, 0},
		},
		{
			name:     "Clear last bit in lo element",
			initial:  Bits128{1 << 31, 0},
			index:    31,
			expected: Bits128{0, 0},
		},
		{
			name:     "Clear first bit in hi element",
			initial:  Bits128{0, 1},
			index:    64,
			expected: Bits128{0, 0},
		},
		{
			name:     "Clear bit in middle of hi element",
			initial:  Bits128{0, 1 << 21},
			index:    85,
			expected: Bits128{0, 0},
		},
		{
			name:     "Clear bit that's already cleared",
			initial:  Bits128{3, 0},
			index:    5,
			expected: Bits128{3, 0},
		},
		{
			name:     "Clear with multiple bits set",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			index:    33,
			expected: Bits128{0xFFFFFFFDFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
		},
		{
			name:     "Clear last possible bit",
			initial:  Bits128{6845, 1 << 63},
			index:    127,
			expected: Bits128{6845, 0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.Clear(tt.index)
			if result != tt.expected {
				t.Errorf("FuseSet.Clear(%d) = %v; want %v", tt.index, result, tt.expected)
			}
		})
	}
}
func TestFuseSet_Diff(t *testing.T) {
	tests := []struct {
		name     string
		f        Bits128
		g        Bits128
		expected Bits128
	}{
		{
			name:     "All zeros",
			f:        Bits128{0, 0},
			g:        Bits128{0, 0},
			expected: Bits128{0, 0},
		},
		{
			name:     "All ones",
			f:        Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			g:        Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			expected: Bits128{0, 0},
		},
		{
			name:     "Alternating bits",
			f:        Bits128{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
			g:        Bits128{0x5555555555555555, 0x5555555555555555},
			expected: Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
		},
		{
			name:     "Single bit difference",
			f:        Bits128{1, 0},
			g:        Bits128{0, 0},
			expected: Bits128{1, 0},
		},
		{
			name:     "Mixed differences",
			f:        Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
			g:        Bits128{0xFEDCBA9876543210, 0x123456789ABCDEF0},
			expected: Bits128{0xECE8ECE0ECE8ECE0, 0xECE8ECE0ECE8ECE0},
		},
		{
			name:     "Partial differences",
			f:        Bits128{0xFFFF00000000FFFF, 0xF0F0F0F00F0F0F0F},
			g:        Bits128{0x0000FFFFFFFF0000, 0x0F0F0F0FF0F0F0F0},
			expected: Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.f.Diff(tt.g)
			if result != tt.expected {
				t.Errorf("FuseSet.Diff() = %v; want %v", result, tt.expected)
			}
		})
	}
}
func TestFuseSet_Union(t *testing.T) {
	tests := []struct {
		name     string
		f        Bits128
		g        Bits128
		expected Bits128
	}{
		{
			name:     "Union of empty sets",
			f:        Bits128{0, 0},
			g:        Bits128{0, 0},
			expected: Bits128{0, 0},
		},
		{
			name:     "Union with self",
			f:        Bits128{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
			g:        Bits128{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
			expected: Bits128{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
		},
		{
			name:     "Union of complementary sets",
			f:        Bits128{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
			g:        Bits128{0x5555555555555555, 0x5555555555555555},
			expected: Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
		},
		{
			name:     "Union with one empty set",
			f:        Bits128{0x12312345678, 0xBEEF9ABCDEF0},
			g:        Bits128{0, 0},
			expected: Bits128{0x12312345678, 0xBEEF9ABCDEF0},
		},
		{
			name:     "Union of partially overlapping sets",
			f:        Bits128{0xF0F0F0F0FF, 0xF0F0F0F000},
			g:        Bits128{0x0F0F0F0FFF, 0x0F0F0F0FF0},
			expected: Bits128{0xFFFFFFFFFF, 0xFFFFFFFFF0},
		},
		{
			name:     "Union of sets with bits in different positions",
			f:        Bits128{0x00FF00FFFF00FF00, 0},
			g:        Bits128{0, 0xFF00FF0000FF00FF},
			expected: Bits128{0x00FF00FFFF00FF00, 0xFF00FF0000FF00FF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.f.Union(tt.g)
			if result != tt.expected {
				t.Errorf("FuseSet.Union() = %v; want %v", result, tt.expected)
			}
		})
	}
}
func TestFuseSet_Intersection(t *testing.T) {
	tests := []struct {
		name     string
		f        Bits128
		g        Bits128
		expected Bits128
	}{
		{
			name:     "Intersection of full sets",
			f:        Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			g:        Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			expected: Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
		},
		{
			name:     "Intersection with empty set",
			f:        Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			g:        Bits128{0, 0},
			expected: Bits128{0, 0},
		},
		{
			name:     "Intersection of disjoint sets",
			f:        Bits128{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
			g:        Bits128{0x5555555555555555, 0x5555555555555555},
			expected: Bits128{0, 0},
		},
		{
			name:     "Intersection of partially overlapping sets",
			f:        Bits128{0xF0F0F0F0F0F0F0F0, 0xF0F0F0F0F0F0F0F0},
			g:        Bits128{0xFF00FF00FF00FF00, 0xFF00FF00FF00FF00},
			expected: Bits128{0xF000F000F000F000, 0xF000F000F000F000},
		},
		{
			name:     "Intersection with self",
			f:        Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
			g:        Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
			expected: Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
		},
		{
			name:     "Intersection with alternating bits",
			f:        Bits128{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
			g:        Bits128{0xCCCCCCCCCCCCCCCC, 0xCCCCCCCCCCCCCCCC},
			expected: Bits128{0x8888888888888888, 0x8888888888888888},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.f.Intersection(tt.g)
			if result != tt.expected {
				t.Errorf("FuseSet.Intersection() = %v; want %v", result, tt.expected)
			}
		})
	}
}
func TestFuseSet_ShiftLeft(t *testing.T) {
	tests := []struct {
		name     string
		initial  Bits128
		shift    uint8
		expected Bits128
	}{
		{
			name:     "Shift by 0",
			initial:  Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
			shift:    0,
			expected: Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
		},
		{
			name:     "Shift by 1",
			initial:  Bits128{0x8000000180000001, 0x8000000180000001},
			shift:    1,
			expected: Bits128{0x0000000300000002, 0x0000000300000003},
		},
		{
			name:     "Shift by 31",
			initial:  Bits128{0x00000000FFFFFFFF, 0x0000000000000000},
			shift:    31,
			expected: Bits128{0x7FFFFFFF80000000, 0x0000000000000000},
		},
		{
			name:     "Shift by 32",
			initial:  Bits128{0x00000000FFFFFFFF, 0x0000000000000000},
			shift:    32,
			expected: Bits128{0xFFFFFFFF00000000, 0x0000000000000000},
		},
		{
			name:     "Shift by 33",
			initial:  Bits128{0x00000000FFFFFFFF, 0x0000000000000000},
			shift:    33,
			expected: Bits128{0xFFFFFFFE00000000, 0x0000000000000001},
		},
		{
			name:     "Shift by 64",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0x0000000000000000},
			shift:    64,
			expected: Bits128{0x0000000000000000, 0xFFFFFFFFFFFFFFFF},
		},
		{
			name:     "Shift by 96",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0xABCDEF0123456789},
			shift:    96,
			expected: Bits128{0x0000000000000000, 0xFFFFFFFF00000000},
		},
		{
			name:     "Shift by 127",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			shift:    127,
			expected: Bits128{0x0000000000000000, 0x8000000000000000},
		},
		{
			name:     "Shift by 128",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			shift:    128,
			expected: Bits128{0x0000000000000000, 0x0000000000000000},
		},
		{
			name:     "Shift by 129",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			shift:    129,
			expected: Bits128{0x0000000000000000, 0x0000000000000000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.ShiftLeft(tt.shift)
			if result != tt.expected {
				t.Errorf("FuseSet.ShiftLeft(%d) = %v; want %v", tt.shift, result, tt.expected)
			}
		})
	}
}
func TestFuseSet_ShiftRight(t *testing.T) {
	tests := []struct {
		name     string
		initial  Bits128
		shift    uint8
		expected Bits128
	}{
		{
			name:     "Shift by 0",
			initial:  Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
			shift:    0,
			expected: Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
		},
		{
			name:     "Shift by 1",
			initial:  Bits128{0x8000000180000001, 0x8000000180000001},
			shift:    1,
			expected: Bits128{0xC0000000C0000000, 0x40000000C0000000},
		},
		{
			name:     "Shift by 31",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			shift:    31,
			expected: Bits128{0xFFFFFFFFFFFFFFFF, 0x00000001FFFFFFFF},
		},
		{
			name:     "Shift by 32",
			initial:  Bits128{0xFFFFFFFF00000000, 0x0000000000000000},
			shift:    32,
			expected: Bits128{0x00000000FFFFFFFF, 0x0000000000000000},
		},
		{
			name:     "Shift by 33",
			initial:  Bits128{0xFFFFFFFF00000000, 0x0000000080000000},
			shift:    33,
			expected: Bits128{0x400000007FFFFFFF, 0x0000000000000000},
		},
		{
			name:     "Shift by 64",
			initial:  Bits128{0x0000000000000000, 0xFFFFFFFFFFFFFFFF},
			shift:    64,
			expected: Bits128{0xFFFFFFFFFFFFFFFF, 0x0000000000000000},
		},
		{
			name:     "Shift by 96",
			initial:  Bits128{0x0000000000000000, 0xFFFFFFFF00000000},
			shift:    96,
			expected: Bits128{0x00000000FFFFFFFF, 0x0000000000000000},
		},
		{
			name:     "Shift by 127",
			initial:  Bits128{0x0000000000000000, 0x8000000000000000},
			shift:    127,
			expected: Bits128{0x0000000000000001, 0x0000000000000000},
		},
		{
			name:     "Shift by 128",
			initial:  Bits128{0x0000000000000000, 0xFFFFFFFF00000000},
			shift:    128,
			expected: Bits128{0x0000000000000000, 0x0000000000000000},
		},
		{
			name:     "Shift by 129",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			shift:    129,
			expected: Bits128{0x0000000000000000, 0x0000000000000000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.ShiftRight(tt.shift)
			if result != tt.expected {
				t.Errorf("FuseSet.ShiftRight(%d) = %v; want %v", tt.shift, result, tt.expected)
			}
		})
	}
}
func TestFuseSet_Invert(t *testing.T) {
	tests := []struct {
		name     string
		initial  Bits128
		expected Bits128
	}{
		{
			name:     "Invert all zeros",
			initial:  Bits128{0x0000000000000000, 0x0000000000000000},
			expected: Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
		},
		{
			name:     "Invert all ones",
			initial:  Bits128{0xFFFFFFFFFFFFFFFF, 0xFFFFFFFFFFFFFFFF},
			expected: Bits128{0x0000000000000000, 0x0000000000000000},
		},
		{
			name:     "Invert alternating bits",
			initial:  Bits128{0xAAAAAAAAAAAAAAAA, 0xAAAAAAAAAAAAAAAA},
			expected: Bits128{0x5555555555555555, 0x5555555555555555},
		},
		{
			name:     "Invert mixed values",
			initial:  Bits128{0x123456789ABCDEF0, 0xFEDCBA9876543210},
			expected: Bits128{0xEDCBA9876543210F, 0x0123456789ABCDEF},
		},
		{
			name:     "Invert other pattern",
			initial:  Bits128{0xFFFFFFFF00000000, 0xFFFFFFFF00000000},
			expected: Bits128{0x00000000FFFFFFFF, 0x00000000FFFFFFFF},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.initial.Invert()
			if result != tt.expected {
				t.Errorf("FuseSet.Invert() = %v; want %v", result, tt.expected)
			}
		})
	}
}

var charSets = []string{
	"0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz:_",
	"0123456789ABCDEF",
	"0123456789",
	"Ee",
	"!\"#$%&'()*+,-./0123456789:;<=>?@ABCDEFGHIJKLMNOPQRSTUVWXYZ[\\]^_`abcdefghijklmnopqrstuvwxyz{|}~",
}

func generateMatch(pattern string, seed int64) string {
	// deterministic is better than flaky
	var s [32]byte
	binary.NativeEndian.PutUint64(s[:], uint64(seed))
	rnd := rand.NewChaCha8(s)
	b := make([]byte, len(pattern))
	for i := 0; i < len(pattern); i++ {
		c := pattern[i]
		if c == '\x7F' {
			if i != len(pattern)-1 {
				panic("broken pattern terminator")
			}
			b[i] = 'X'
			return string(b) + "XX"
		}
		if c >= 32 {
			b[i] = pattern[i]
			continue
		}
		if int(c) >= len(charSets) {
			panic("broken pattern matchset")
		}
		switch seed {
		case 0:
			b[i] = charSets[c][0]
		case -1:
			b[i] = charSets[c][len(charSets[c])-1]
		default:
			// good enough ¯\_(ツ)_/¯
			b[i] = charSets[c][rnd.Uint64()%uint64(len(charSets[c]))]
		}
	}
	return string(b)
}
