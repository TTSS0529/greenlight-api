package data

import (
	"errors"
	"testing"
)

func TestRuntimeMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		runtime  Runtime
		expected string
	}{
		{
			name:     "90 minutes",
			runtime:  Runtime(90),
			expected: `"90 mins"`,
		},
		{
			name:     "0 minutes",
			runtime:  Runtime(0),
			expected: `"0 mins"`,
		},
		{
			name:     "negative minutes",
			runtime:  Runtime(-10),
			expected: `"-10 mins"`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.runtime.MarshalJSON()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if string(got) != tt.expected {
				t.Errorf("expected %s, got %s", tt.expected, got)
			}
		})
	}
}

func TestRuntimeUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    Runtime
		expectError error
	}{
		{
			name:     "valid runtime",
			input:    `"90 mins"`,
			expected: Runtime(90),
		},
		{
			name:     "zero runtime",
			input:    `"0 mins"`,
			expected: Runtime(0),
		},
		{
			name:        "missing mins suffix",
			input:       `"90"`,
			expectError: ErrInvalidRuntimeFormat,
		},
		{
			name:        "wrong suffix",
			input:       `"90 minutes"`,
			expectError: ErrInvalidRuntimeFormat,
		},
		{
			name:        "invalid number",
			input:       `"abc mins"`,
			expectError: ErrInvalidRuntimeFormat,
		},
		{
			name:        "invalid json string",
			input:       `90 mins`,
			expectError: ErrInvalidRuntimeFormat,
		},
		{
			name:        "too many parts",
			input:       `"90 mins extra"`,
			expectError: ErrInvalidRuntimeFormat,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got Runtime
			err := got.UnmarshalJSON([]byte(tt.input))
			if !errors.Is(err, tt.expectError) {
				t.Fatalf("expected error %v, got %v", tt.expectError, err)
			}
			if err == nil && got != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, got)
			}
		})
	}
}
