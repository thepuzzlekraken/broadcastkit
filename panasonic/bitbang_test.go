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
func TestFuseSet_Set(t *testing.T) {
	tests := []struct {
		name     string
		initial  FuseSet
		index    uint8
		expected FuseSet
	}{
		{
			name:     "Set first bit",
			initial:  FuseSet{0, 0, 0, 0},
			index:    0,
			expected: FuseSet{1, 0, 0, 0},
		},
		{
			name:     "Set last bit in first element",
			initial:  FuseSet{0, 0, 0, 0},
			index:    31,
			expected: FuseSet{1 << 31, 0, 0, 0},
		},
		{
			name:     "Set first bit in second element",
			initial:  FuseSet{0, 0, 0, 0},
			index:    32,
			expected: FuseSet{0, 1, 0, 0},
		},
		{
			name:     "Set bit in middle of third element",
			initial:  FuseSet{0, 0, 0, 0},
			index:    85,
			expected: FuseSet{0, 0, 1 << 21, 0},
		},
		{
			name:     "Set bit that's already set",
			initial:  FuseSet{1 << 5, 0, 0, 0},
			index:    5,
			expected: FuseSet{1 << 5, 0, 0, 0},
		},
		{
			name:     "Set with multiple bits",
			initial:  FuseSet{1, 0, 1 << 15, 0},
			index:    33,
			expected: FuseSet{1, 2, 1 << 15, 0},
		},
		{
			name:     "Set last possible bit",
			initial:  FuseSet{0, 0, 0, 0},
			index:    127,
			expected: FuseSet{0, 0, 0, 1 << 31},
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
		initial  FuseSet
		index    uint8
		expected FuseSet
	}{
		{
			name:     "Clear first bit",
			initial:  FuseSet{1, 0, 0, 0},
			index:    0,
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "Clear last bit in first element",
			initial:  FuseSet{1 << 31, 0, 0, 0},
			index:    31,
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "Clear first bit in second element",
			initial:  FuseSet{0, 1, 0, 0},
			index:    32,
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "Clear bit in middle of third element",
			initial:  FuseSet{0, 0, 1 << 21, 0},
			index:    85,
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "Clear bit that's already cleared",
			initial:  FuseSet{3, 0, 0, 0},
			index:    5,
			expected: FuseSet{3, 0, 0, 0},
		},
		{
			name:     "Clear with multiple bits set",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			index:    33,
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFD, 0xFFFFFFFF, 0xFFFFFFFF},
		},
		{
			name:     "Clear last possible bit",
			initial:  FuseSet{0, 1, 0, 1 << 31},
			index:    127,
			expected: FuseSet{0, 1, 0, 0},
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
		f        FuseSet
		g        FuseSet
		expected FuseSet
	}{
		{
			name:     "All zeros",
			f:        FuseSet{0, 0, 0, 0},
			g:        FuseSet{0, 0, 0, 0},
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "All ones",
			f:        FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			g:        FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "Alternating bits",
			f:        FuseSet{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
			g:        FuseSet{0x55555555, 0x55555555, 0x55555555, 0x55555555},
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
		},
		{
			name:     "Single bit difference",
			f:        FuseSet{1, 0, 0, 0},
			g:        FuseSet{0, 0, 0, 0},
			expected: FuseSet{1, 0, 0, 0},
		},
		{
			name:     "Mixed differences",
			f:        FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
			g:        FuseSet{0xFEDCBA98, 0x76543210, 0x12345678, 0x9ABCDEF0},
			expected: FuseSet{0xECE8ECE0, 0xECE8ECE0, 0xECE8ECE0, 0xECE8ECE0},
		},
		{
			name:     "Partial differences",
			f:        FuseSet{0xFFFF0000, 0x0000FFFF, 0xF0F0F0F0, 0x0F0F0F0F},
			g:        FuseSet{0x0000FFFF, 0xFFFF0000, 0x0F0F0F0F, 0xF0F0F0F0},
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
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
		f        FuseSet
		g        FuseSet
		expected FuseSet
	}{
		{
			name:     "Union of empty sets",
			f:        FuseSet{0, 0, 0, 0},
			g:        FuseSet{0, 0, 0, 0},
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "Union with self",
			f:        FuseSet{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
			g:        FuseSet{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
			expected: FuseSet{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
		},
		{
			name:     "Union of complementary sets",
			f:        FuseSet{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
			g:        FuseSet{0x55555555, 0x55555555, 0x55555555, 0x55555555},
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
		},
		{
			name:     "Union with one empty set",
			f:        FuseSet{0x12345678, 0x9ABCDEF0, 0, 0},
			g:        FuseSet{0, 0, 0, 0},
			expected: FuseSet{0x12345678, 0x9ABCDEF0, 0, 0},
		},
		{
			name:     "Union of partially overlapping sets",
			f:        FuseSet{0xF0F0F0F0, 0xF0F0F0F0, 0, 0},
			g:        FuseSet{0x0F0F0F0F, 0x0F0F0F0F, 0, 0},
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0, 0},
		},
		{
			name:     "Union of sets with bits in different positions",
			f:        FuseSet{0x00FF00FF, 0, 0xFF00FF00, 0},
			g:        FuseSet{0, 0xFF00FF00, 0, 0x00FF00FF},
			expected: FuseSet{0x00FF00FF, 0xFF00FF00, 0xFF00FF00, 0x00FF00FF},
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
		f        FuseSet
		g        FuseSet
		expected FuseSet
	}{
		{
			name:     "Intersection of full sets",
			f:        FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			g:        FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
		},
		{
			name:     "Intersection with empty set",
			f:        FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			g:        FuseSet{0, 0, 0, 0},
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "Intersection of disjoint sets",
			f:        FuseSet{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
			g:        FuseSet{0x55555555, 0x55555555, 0x55555555, 0x55555555},
			expected: FuseSet{0, 0, 0, 0},
		},
		{
			name:     "Intersection of partially overlapping sets",
			f:        FuseSet{0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0, 0xF0F0F0F0},
			g:        FuseSet{0xFF00FF00, 0xFF00FF00, 0xFF00FF00, 0xFF00FF00},
			expected: FuseSet{0xF000F000, 0xF000F000, 0xF000F000, 0xF000F000},
		},
		{
			name:     "Intersection with self",
			f:        FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
			g:        FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
			expected: FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
		},
		{
			name:     "Intersection with alternating bits",
			f:        FuseSet{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
			g:        FuseSet{0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC, 0xCCCCCCCC},
			expected: FuseSet{0x88888888, 0x88888888, 0x88888888, 0x88888888},
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
		initial  FuseSet
		shift    uint
		expected FuseSet
	}{
		{
			name:     "Shift by 0",
			initial:  FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
			shift:    0,
			expected: FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
		},
		{
			name:     "Shift by 1",
			initial:  FuseSet{0x80000001, 0x80000001, 0x80000001, 0x80000001},
			shift:    1,
			expected: FuseSet{0x00000002, 0x00000003, 0x00000003, 0x00000003},
		},
		{
			name:     "Shift by 31",
			initial:  FuseSet{0xFFFFFFFF, 0x00000000, 0x00000000, 0x00000000},
			shift:    31,
			expected: FuseSet{0x80000000, 0x7FFFFFFF, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 32",
			initial:  FuseSet{0xFFFFFFFF, 0x00000000, 0x00000000, 0x00000000},
			shift:    32,
			expected: FuseSet{0x00000000, 0xFFFFFFFF, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 33",
			initial:  FuseSet{0xFFFFFFFF, 0x00000000, 0x00000000, 0x00000000},
			shift:    33,
			expected: FuseSet{0x00000000, 0xFFFFFFFE, 0x00000001, 0x00000000},
		},
		{
			name:     "Shift by 64",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0x00000000, 0x00000000},
			shift:    64,
			expected: FuseSet{0x00000000, 0x00000000, 0xFFFFFFFF, 0xFFFFFFFF},
		},
		{
			name:     "Shift by 96",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0x00000000},
			shift:    96,
			expected: FuseSet{0x00000000, 0x00000000, 0x00000000, 0xFFFFFFFF},
		},
		{
			name:     "Shift by 127",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			shift:    127,
			expected: FuseSet{0x00000000, 0x00000000, 0x00000000, 0x80000000},
		},
		{
			name:     "Shift by 128",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			shift:    128,
			expected: FuseSet{0x00000000, 0x00000000, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 129",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			shift:    129,
			expected: FuseSet{0x00000000, 0x00000000, 0x00000000, 0x00000000},
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
		initial  FuseSet
		shift    uint
		expected FuseSet
	}{
		{
			name:     "Shift by 0",
			initial:  FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
			shift:    0,
			expected: FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
		},
		{
			name:     "Shift by 1",
			initial:  FuseSet{0x80000001, 0x80000001, 0x80000001, 0x80000001},
			shift:    1,
			expected: FuseSet{0xC0000000, 0xC0000000, 0xC0000000, 0x40000000},
		},
		{
			name:     "Shift by 31",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			shift:    31,
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0x00000001},
		},
		{
			name:     "Shift by 32",
			initial:  FuseSet{0x00000000, 0xFFFFFFFF, 0x00000000, 0x00000000},
			shift:    32,
			expected: FuseSet{0xFFFFFFFF, 0x00000000, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 33",
			initial:  FuseSet{0x00000000, 0xFFFFFFFF, 0x80000000, 0x00000000},
			shift:    33,
			expected: FuseSet{0x7FFFFFFF, 0x40000000, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 64",
			initial:  FuseSet{0x00000000, 0x00000000, 0xFFFFFFFF, 0xFFFFFFFF},
			shift:    64,
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 96",
			initial:  FuseSet{0x00000000, 0x00000000, 0x00000000, 0xFFFFFFFF},
			shift:    96,
			expected: FuseSet{0xFFFFFFFF, 0x00000000, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 127",
			initial:  FuseSet{0x00000000, 0x00000000, 0x00000000, 0x80000000},
			shift:    127,
			expected: FuseSet{0x00000001, 0x00000000, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 128",
			initial:  FuseSet{0x00000000, 0x00000000, 0x00000000, 0xFFFFFFFF},
			shift:    128,
			expected: FuseSet{0x00000000, 0x00000000, 0x00000000, 0x00000000},
		},
		{
			name:     "Shift by 129",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			shift:    129,
			expected: FuseSet{0x00000000, 0x00000000, 0x00000000, 0x00000000},
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
		initial  FuseSet
		expected FuseSet
	}{
		{
			name:     "Invert all zeros",
			initial:  FuseSet{0x00000000, 0x00000000, 0x00000000, 0x00000000},
			expected: FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
		},
		{
			name:     "Invert all ones",
			initial:  FuseSet{0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
			expected: FuseSet{0x00000000, 0x00000000, 0x00000000, 0x00000000},
		},
		{
			name:     "Invert alternating bits",
			initial:  FuseSet{0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA, 0xAAAAAAAA},
			expected: FuseSet{0x55555555, 0x55555555, 0x55555555, 0x55555555},
		},
		{
			name:     "Invert mixed values",
			initial:  FuseSet{0x12345678, 0x9ABCDEF0, 0xFEDCBA98, 0x76543210},
			expected: FuseSet{0xEDCBA987, 0x6543210F, 0x01234567, 0x89ABCDEF},
		},
		{
			name:     "Invert other pattern",
			initial:  FuseSet{0xFFFFFFFF, 0x00000000, 0xFFFFFFFF, 0x00000000},
			expected: FuseSet{0x00000000, 0xFFFFFFFF, 0x00000000, 0xFFFFFFFF},
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
