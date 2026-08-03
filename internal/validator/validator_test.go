package validator

import (
	"regexp"
	"testing"
)

func TestNew(t *testing.T) {
	v := New()
	if v == nil {
		t.Fatal("expected validator, got nil")
	}
	if v.Errors == nil {
		t.Fatal("expected initialized Errors map")
	}
	if !v.Valid() {
		t.Error("new validator should be valid")
	}
}

func TestValidator_AddError(t *testing.T) {
	v := New()
	v.AddError("name", "must be provided")
	if got := v.Errors["name"]; got != "must be provided" {
		t.Errorf("got %q; want %q", got, "must be provided")
	}
	// Existing error should not be overwritten.
	v.AddError("name", "another error")
	if got := v.Errors["name"]; got != "must be provided" {
		t.Errorf("got %q; want %q", got, "must be provided")
	}
}

func TestValidator_Check(t *testing.T) {
	tests := []struct {
		name      string
		ok        bool
		wantValid bool
	}{
		{
			name:      "valid",
			ok:        true,
			wantValid: true,
		},
		{
			name:      "invalid",
			ok:        false,
			wantValid: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := New()
			v.Check(tt.ok, "field", "error")
			if got := v.Valid(); got != tt.wantValid {
				t.Errorf("got %v; want %v", got, tt.wantValid)
			}
		})
	}
}

func TestValidator_Valid(t *testing.T) {
	v := New()
	if !v.Valid() {
		t.Error("expected validator to be valid")
	}
	v.AddError("field", "error")
	if v.Valid() {
		t.Error("expected validator to be invalid")
	}
}

func TestIn(t *testing.T) {
	tests := []struct {
		name  string
		value string
		list  []string
		want  bool
	}{
		{
			name:  "value exists",
			value: "go",
			list:  []string{"c", "go", "cpp"},
			want:  true,
		},
		{
			name:  "value missing",
			value: "rust",
			list:  []string{"c", "go", "cpp"},
			want:  false,
		},
		{
			name:  "empty list",
			value: "go",
			list:  nil,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := In(tt.value, tt.list...)
			if got != tt.want {
				t.Errorf("got %v; want %v", got, tt.want)
			}
		})
	}
}

func TestMatches(t *testing.T) {
	tests := []struct {
		name  string
		value string
		rx    *regexp.Regexp
		want  bool
	}{
		{
			name:  "matches",
			value: "12345",
			rx:    regexp.MustCompile(`^[0-9]+$`),
			want:  true,
		},
		{
			name:  "does not match",
			value: "abc123",
			rx:    regexp.MustCompile(`^[0-9]+$`),
			want:  false,
		},
		{
			name:  "email",
			value: "alice@example.com",
			rx:    EmailRX,
			want:  true,
		},
		{
			name:  "invalid email",
			value: "alice@@example.com",
			rx:    EmailRX,
			want:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Matches(tt.value, tt.rx)
			if got != tt.want {
				t.Errorf("got %v; want %v", got, tt.want)
			}
		})
	}
}

func TestUnique(t *testing.T) {
	tests := []struct {
		name   string
		values []string
		want   bool
	}{
		{
			name:   "all unique",
			values: []string{"go", "cpp", "rust"},
			want:   true,
		},
		{
			name:   "duplicate",
			values: []string{"go", "cpp", "go"},
			want:   false,
		},
		{
			name:   "empty",
			values: nil,
			want:   true,
		},
		{
			name:   "single",
			values: []string{"go"},
			want:   true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Unique(tt.values)
			if got != tt.want {
				t.Errorf("got %v; want %v", got, tt.want)
			}
		})
	}
}
