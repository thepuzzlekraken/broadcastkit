package panasonic

import (
	"testing"
)

func TestAWErrorResponseSignature(t *testing.T) {
	tests := []struct {
		name     string
		flag     string
		wantType awHint
		wantSig  string
	}{
		{
			name:     "Empty flag",
			flag:     "",
			wantType: awPtz | awCam,
			wantSig:  "\x03R\x01:",
		},
		{
			name:     "Short flag",
			flag:     " ",
			wantType: awPtz | awCam,
			wantSig:  "\x03R\x01:\x00",
		},
		{
			name:     "Medium flag",
			flag:     "AB",
			wantType: awPtz | awCam,
			wantSig:  "\x03R\x01:\x00\x00",
		},
		{
			name:     "Max length flag",
			flag:     "   ",
			wantType: awPtz | awCam,
			wantSig:  "\x03R\x01:\x00\x00\x00",
		},
		{
			name:     "Oversized flag",
			flag:     "TOOLONG",
			wantType: awPtz | awCam,
			wantSig:  "\x03R\x01:\x00\x00\x00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &AWError{Flag: tt.flag}
			gotType, gotSig := e.responseSignature()
			if gotType != tt.wantType {
				t.Errorf("AWError.responseSignature() type = %v, want %v", gotType, tt.wantType)
			}
			if gotSig != tt.wantSig {
				t.Errorf("AWError.responseSignature() sig = %v, want %v", gotSig, tt.wantSig)
			}
		})
	}
}
