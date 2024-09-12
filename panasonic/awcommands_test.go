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
				Offset: 0,
				Bits:   Bits64(0).Set(0).Set(39),
			},
		},
		{
			name: "Offset 1",
			cmd:  "pE018000000010",
			expected: AWPresetEntries{
				Offset: 1,
				Bits:   Bits64(0).Set(4).Set(39),
			},
		},
		{
			name: "Offset 2",
			cmd:  "pE020000080100",
			expected: AWPresetEntries{
				Offset: 2,
				Bits:   Bits64(0).Set(8).Set(19),
			},
		},
		{
			name: "All Ones",
			cmd:  "pE00FFFFFFFFFF",
			expected: AWPresetEntries{
				Offset: 0,
				Bits:   Bits64(0xFFFFFFFFFF),
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
				Offset: 0,
				Bits:   Bits64(0).Set(0).Set(5).Set(10),
			},
			expected: "pE000000000421",
		},
		{
			name: "Offset 1 with some entries",
			input: AWPresetEntries{
				Offset: 1,
				Bits:   Bits64(0).Set(0).Set(5).Set(10).Set(39),
			},
			expected: "pE018000000421",
		},
		{
			name: "Offset 2 with some entries",
			input: AWPresetEntries{
				Offset: 2,
				Bits:   Bits64(0).Set(0).Set(5).Set(10),
			},
			expected: "pE020000000421",
		},
		{
			name: "Invalid negative offset",
			input: AWPresetEntries{
				Offset: -1,
				Bits:   Bits64(0),
			},
			expected: "pEFF0000000000",
		},
		{
			name: "Empty entries",
			input: AWPresetEntries{
				Offset: 0,
				Bits:   Bits64(0),
			},
			expected: "pE000000000000",
		},
		{
			name: "All entries set",
			input: AWPresetEntries{
				Offset: 0,
				Bits:   Bits64(0xFFFFFFFFFF),
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
		offset   Offset
		expected Bits128
	}{
		{
			name:     "Offset 0",
			offset:   0,
			expected: Bits128{0xFFFFFFFFFF, 0x0},
		},
		{
			name:     "Offset 1",
			offset:   1,
			expected: Bits128{0xFFFFFF0000000000, 0xFFFF},
		},
		{
			name:     "Offset 2",
			offset:   2,
			expected: Bits128{0x0, 0xFFFFFFFFFF0000},
		},
		{
			name:     "Offset 3",
			offset:   3,
			expected: Bits128{0x0, 0xFF00000000000000},
		},
		{
			name:     "Negative Offset",
			offset:   -1,
			expected: Bits128{0x0, 0x0},
		},
		{
			name:     "Large Offset",
			offset:   10,
			expected: Bits128{0x0, 0x0},
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
