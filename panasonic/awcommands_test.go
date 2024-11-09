package panasonic

import "testing"

func TestAWPresetEntries(t *testing.T) {
	tests := []struct {
		name   string
		input  Bits128
		want00 AWPresetEntries00
		want01 AWPresetEntries01
		want02 AWPresetEntries02
	}{
		{
			name:   "zero bits",
			input:  Bits128{},
			want00: AWPresetEntries00{Bits: 0},
			want01: AWPresetEntries01{Bits: 0},
			want02: AWPresetEntries02{Bits: 0},
		},
		{
			name:   "all bits set",
			input:  Bits128{Hi: 0x00FFFFFFFFFFFFFF, Lo: 0xFFFFFFFFFFFFFFFF},
			want00: AWPresetEntries00{Bits: 0xFFFFFFFFFF},
			want01: AWPresetEntries01{Bits: 0xFFFFFFFFFF},
			want02: AWPresetEntries02{Bits: 0xFFFFFFFFFF},
		},
		{
			name:   "sequential bits",
			input:  Bits128{Hi: 0x003456789ABCDEF0, Lo: 0xFEDCBA9876543210},
			want00: AWPresetEntries00{Bits: 0x9876543210},
			want01: AWPresetEntries01{Bits: 0xDEF0FEDCBA},
			want02: AWPresetEntries02{Bits: 0x3456789ABC},
		},
		{
			name:   "single bit set",
			input:  Bits128{Hi: 0x0000000000000001, Lo: 0x0000000000000000},
			want00: AWPresetEntries00{Bits: 0x0000000000},
			want01: AWPresetEntries01{Bits: 0x0001000000},
			want02: AWPresetEntries02{Bits: 0x0000000000},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got00, got01, got02 := AWPresetEntries(tt.input)
			if got00 != tt.want00 {
				t.Errorf("AWPresetEntries() got00 = %v, want %v", got00, tt.want00)
			}
			if got01 != tt.want01 {
				t.Errorf("AWPresetEntries() got01 = %v, want %v", got01, tt.want01)
			}
			if got02 != tt.want02 {
				t.Errorf("AWPresetEntries() got02 = %v, want %v", got02, tt.want02)
			}
		})
	}
}

func TestAWPresetEntriesMask(t *testing.T) {
	tests := []struct {
		name string
		cmd  interface{ Mask() Bits128 }
		want Bits128
	}{
		{
			name: "AWPresetEntries00",
			cmd:  AWPresetEntries00{Bits: 0xFFFFFFFFFF},
			want: Bits128{0xFFFFFFFFFF, 0x0},
		},
		{
			name: "AWPresetEntries01",
			cmd:  AWPresetEntries01{Bits: 0xFFFFFFFFFF},
			want: Bits128{0xFFFFFF0000000000, 0xFFFF},
		},
		{
			name: "AWPresetEntries02",
			cmd:  AWPresetEntries02{Bits: 0xFFFFFFFFFF},
			want: Bits128{0x0, 0xFFFFFFFFFF0000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.Mask(); got != tt.want {
				t.Errorf("Mask() = %v, want %v", got, tt.want)
			}
		})
	}
}
func TestAWPresetEntriesPresetBits(t *testing.T) {
	tests := []struct {
		name string
		cmd  interface{ PresetBits() Bits128 }
		want Bits128
	}{
		{
			name: "AWPresetEntries00",
			cmd:  AWPresetEntries00{Bits: 0x3214567890},
			want: Bits128{0x3214567890, 0x0},
		},
		{
			name: "AWPresetEntries01",
			cmd:  AWPresetEntries01{Bits: 0xDEADBEEF99},
			want: Bits128{0xBEEF990000000000, 0xDEAD},
		},
		{
			name: "AWPresetEntries02",
			cmd:  AWPresetEntries02{Bits: 0xB000B},
			want: Bits128{0x0, 0xB000B0000},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cmd.PresetBits(); got != tt.want {
				t.Errorf("PresetBits() = %v, want %v", got, tt.want)
			}
		})
	}
}
