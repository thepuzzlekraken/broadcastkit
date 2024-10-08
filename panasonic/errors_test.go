package panasonic

import (
	"testing"
)

func TestAWErrorResponseSignature(t *testing.T) {
	tests := []struct {
		name    string
		flag    string
		wantSig string
	}{
		{
			name:    "Empty flag",
			flag:    "",
			wantSig: "\x03R\x02:",
		},
		{
			name:    "Short flag",
			flag:    " ",
			wantSig: "\x03R\x02:\x7F",
		},
		{
			name:    "Medium flag",
			flag:    "AB",
			wantSig: "\x03R\x02:\x7F",
		},
		{
			name:    "Max length flag",
			flag:    "   ",
			wantSig: "\x03R\x02:\x7F",
		},
		{
			name:    "Oversized flag",
			flag:    "TOOLONG",
			wantSig: "\x03R\x02:\x7F",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AWError{Flag: tt.flag}
			gotSig := e.responseSignature()
			if gotSig != tt.wantSig {
				t.Errorf("AWError.responseSignature() sig = %#v, want %#v", gotSig, tt.wantSig)
			}
		})
	}
}
