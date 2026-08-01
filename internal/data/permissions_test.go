package data

import (
	"testing"
)

func TestPermissions_Include(t *testing.T) {
	tests := []struct {
		name     string
		perms    Permissions
		code     string
		expected bool
	}{
		{
			name:     "permission exists",
			perms:    Permissions{"movies:read", "movies:write"},
			code:     "movies:read",
			expected: true,
		},
		{
			name:     "permission does not exist",
			perms:    Permissions{"movies:read"},
			code:     "admin",
			expected: false,
		},
		{
			name:     "empty permissions",
			perms:    Permissions{},
			code:     "admin",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.perms.Include(tt.code)
			if got != tt.expected {
				t.Fatalf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
