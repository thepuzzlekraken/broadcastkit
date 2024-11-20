package sony

import (
	"testing"
)

func TestHexDecode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"empty string", "", 0, false},
		{"single digit", "5", 5, false},
		{"multiple digits", "123", 291, false},
		{"hex letters", "abc", 2748, false},
		{"mixed case", "AbC", 0, true},
		{"invalid char", "12g3", 0, true},
		{"max value", "fffff", 1048575, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hexDecode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("hexDecode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("hexDecode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHex20Decoder(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"wrong length", "1234", 0, true},
		{"zero value", "00000", 0, false},
		{"positive value", "00123", 291, false},
		{"negative value", "80000", -524288, false},
		{"max positive", "7ffff", 524287, false},
		{"min negative", "80000", -524288, false},
		{"sony max", "09ca7", 40103, false},
		{"sony min", "f6359", -40103, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := hex20Decoder(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("hex20Decoder() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("hex20Decoder() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHexEncoder(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"zero", 0, "00000"},
		{"small number", 291, "00123"},
		{"large number", 2748, "00abc"},
		{"negative", -524288, "80000"},
		{"max positive", 524287, "7ffff"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hexEncoder(tt.input); got[len(got)-5:] != tt.want {
				t.Errorf("hexEncoder() = %v, want %v", got[len(got)-5:], tt.want)
			}
		})
	}
}

func TestHex20Encoder(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  string
	}{
		{"zero", 0, "00000"},
		{"small number", 291, "00123"},
		{"large number", 2748, "00abc"},
		{"negative", -524288, "80000"},
		{"max positive", 524287, "7ffff"},
		{"sony max", 40103, "09ca7"},
		{"sony min", -40103, "f6359"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hex20Encoder(tt.input); got != tt.want {
				t.Errorf("hex20Encoder() = %v, want %v", got, tt.want)
			}
		})
	}
}
