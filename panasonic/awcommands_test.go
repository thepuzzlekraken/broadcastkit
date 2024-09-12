package panasonic

import (
	"reflect"
	"testing"
)

func TestAWPresetEntriesUnpackResponse(t *testing.T) {
	tests := []struct {
		name     string
		cmd      string
		expected AWPresetEntries
	}{
		{
			name: "Offset 0",
			cmd:  "pE008000000001",
			expected: AWPresetEntries{
				Offset:  0,
				Entries: FuseSet{}.Set(0).Set(39),
			},
		},
		{
			name: "Offset 1",
			cmd:  "pE018000000001",
			expected: AWPresetEntries{
				Offset:  1,
				Entries: FuseSet{}.Set(40).Set(79),
			},
		},
		{
			name: "Offset 2",
			cmd:  "pE020000080001",
			expected: AWPresetEntries{
				Offset:  2,
				Entries: FuseSet{}.Set(80).Set(99),
			},
		},
		{
			name: "All Ones",
			cmd:  "pE00FFFFFFFFFF",
			expected: AWPresetEntries{
				Offset:  0,
				Entries: FuseSet{0xFFFFFFFF, 0xFF, 0x0, 0x0},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AWPresetEntries{}
			a.unpackResponse(tt.cmd)
			if !reflect.DeepEqual(a, &tt.expected) {
				t.Errorf("unpackResponse() = %v, want %v", a, tt.expected)
			}
		})
	}
}
func TestAWPresetEntriesPackResponse(t *testing.T) {
	tests := []struct {
		name     string
		input    AWPresetEntries
		expected string
	}{
		{
			name: "Offset 0 with some entries",
			input: AWPresetEntries{
				Offset:  0,
				Entries: FuseSet{}.Set(0).Set(5).Set(10),
			},
			expected: "pE000000000421",
		},
		{
			name: "Offset 1 with some entries",
			input: AWPresetEntries{
				Offset:  1,
				Entries: FuseSet{}.Set(40).Set(45).Set(50),
			},
			expected: "pE010000000421",
		},
		{
			name: "Offset 2 with some entries",
			input: AWPresetEntries{
				Offset:  2,
				Entries: FuseSet{}.Set(80).Set(85).Set(90),
			},
			expected: "pE020000000421",
		},
		{
			name: "Invalid negative offset",
			input: AWPresetEntries{
				Offset:  -1,
				Entries: FuseSet{},
			},
			expected: "pEFF0000000000",
		},
		{
			name: "Empty entries",
			input: AWPresetEntries{
				Offset:  0,
				Entries: FuseSet{},
			},
			expected: "pE000000000000",
		},
		{
			name: "All entries set",
			input: AWPresetEntries{
				Offset:  0,
				Entries: FuseSet{0xFFFFFFFF, 0xFF, 0x0, 0x0},
			},
			expected: "pE00FFFFFFFFFF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.input.packResponse()
			if result != tt.expected {
				t.Errorf("packResponse() = %v, want %v", result, tt.expected)
			}
		})
	}
}
func TestFuseOffset_Mask(t *testing.T) {
	tests := []struct {
		name     string
		offset   FuseOffset
		expected FuseSet
	}{
		{
			name:     "Offset 0",
			offset:   0,
			expected: FuseSet{0xFFFFFFFF, 0xFF, 0x0, 0x0},
		},
		{
			name:     "Offset 1",
			offset:   1,
			expected: FuseSet{0x0, 0xFFFFFF00, 0xFFFF, 0x0},
		},
		{
			name:     "Offset 2",
			offset:   2,
			expected: FuseSet{0x0, 0x0, 0xFFFF0000, 0xFFFFFF},
		},
		{
			name:     "Offset 3",
			offset:   3,
			expected: FuseSet{0x0, 0x0, 0x0, 0xFF000000},
		},
		{
			name:     "Negative Offset",
			offset:   -1,
			expected: FuseSet{0x0, 0x0, 0x0, 0x0},
		},
		{
			name:     "Large Offset",
			offset:   10,
			expected: FuseSet{0x0, 0x0, 0x0, 0x0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.offset.Mask()
			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Offset(%d).Mask() = %x, want %x", tt.offset, result, tt.expected)
			}
		})
	}
}
