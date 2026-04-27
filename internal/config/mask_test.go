package config

import (
	"testing"
)

func TestMaskKey(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "TestMaskKey_LongToken",
			input:    "sk-ant-api03-abc1",
			expected: "sk-ant-...abc1",
		},
		{
			name:     "TestMaskKey_ShortToken",
			input:    "short",
			expected: "****",
		},
		{
			name:     "TestMaskKey_ExactlyElevenChars",
			input:    "12345678901",
			expected: "1234567...8901",
		},
		{
			name:     "TestMaskKey_EmptyString",
			input:    "",
			expected: "<empty>",
		},
		{
			name:     "TestMaskKey_ExactlySevenChars",
			input:    "1234567",
			expected: "****",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MaskKey(tt.input)
			if result != tt.expected {
				t.Errorf("MaskKey(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}