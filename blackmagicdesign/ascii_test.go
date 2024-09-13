package blackmagicdesign

import (
	"bytes"
	"testing"
)

func TestUppercase(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "all uppercase",
			input:    []byte("HELLO WORLD"),
			expected: []byte("HELLO WORLD"),
		},
		{
			name:     "mixed case",
			input:    []byte("HeLLo WoRLd"),
			expected: []byte("HELLO WORLD"),
		},
		{
			name:     "all lowercase",
			input:    []byte("hello world"),
			expected: []byte("HELLO WORLD"),
		},
		{
			name:     "with numbers and symbols",
			input:    []byte("h3ll0 w0rld!"),
			expected: []byte("H3LL0 W0RLD!"),
		},
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "non-ascii characters",
			input:    []byte("héllô wørld"),
			expected: []byte("HéLLô WøRLD"),
		},
		{
			name:     "single character",
			input:    []byte("a"),
			expected: []byte("A"),
		},
		{
			name:     "only numbers and symbols",
			input:    []byte("123!@#"),
			expected: []byte("123!@#"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uppercase(tt.input)
			if !bytes.Equal(tt.input, tt.expected) {
				t.Errorf("uppercase() = %v, want %v", string(tt.input), string(tt.expected))
			}
		})
	}
}
func TestLowercase(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected []byte
	}{
		{
			name:     "all uppercase",
			input:    []byte("HELLO WORLD"),
			expected: []byte("hello world"),
		},
		{
			name:     "mixed case",
			input:    []byte("HeLLo WoRLd"),
			expected: []byte("hello world"),
		},
		{
			name:     "all lowercase",
			input:    []byte("hello world"),
			expected: []byte("hello world"),
		},
		{
			name:     "with numbers and symbols",
			input:    []byte("H3LL0 W0RLD!"),
			expected: []byte("h3ll0 w0rld!"),
		},
		{
			name:     "empty slice",
			input:    []byte{},
			expected: []byte{},
		},
		{
			name:     "non-ascii characters",
			input:    []byte("HÉLLÔ WØRLD"),
			expected: []byte("hÉllÔ wØrld"),
		},
		{
			name:     "single character",
			input:    []byte("A"),
			expected: []byte("a"),
		},
		{
			name:     "only numbers and symbols",
			input:    []byte("123!@#"),
			expected: []byte("123!@#"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lowercase(tt.input)
			if !bytes.Equal(tt.input, tt.expected) {
				t.Errorf("lowercase() = %v, want %v", string(tt.input), string(tt.expected))
			}
		})
	}
}
