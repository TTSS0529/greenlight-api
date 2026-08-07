package data

import (
	"strings"
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/validator"
)

func newTestUser(t *testing.T) *User {
	t.Helper()
	var p password
	err := p.Set("password123")
	if err != nil {
		t.Fatal(err)
	}
	return &User{
		Name:     "Alice",
		Email:    "alice@example.com",
		Password: p,
	}
}

func TestPassword_SetAndMatches(t *testing.T) {
	var p password
	err := p.Set("password123")
	if err != nil {
		t.Fatal(err)
	}
	if p.hash == nil {
		t.Fatal("expected password hash")
	}
	if p.plaintext == nil {
		t.Fatal("expected plaintext")
	}
	ok, err := p.Matches("password123")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Error("expected password to match")
	}
	ok, err = p.Matches("wrongpassword")
	if err != nil {
		t.Fatal(err)
	}
	if ok {
		t.Error("expected password mismatch")
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name  string
		email string
		valid bool
	}{
		{"valid", "abc@test.com", true},
		{"empty", "", false},
		{"invalid", "abc@", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateEmail(v, tt.email)
			if got := v.Valid(); got != tt.valid {
				t.Errorf("got %v want %v", got, tt.valid)
			}
		})
	}
}

func TestValidatePasswordPlaintext(t *testing.T) {
	tests := []struct {
		name     string
		password string
		valid    bool
	}{
		{"valid", "password123", true},
		{"empty", "", false},
		{"short", "1234567", false},
		{"too long", string(make([]byte, 73)), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidatePasswordPlaintext(v, tt.password)
			if got := v.Valid(); got != tt.valid {
				t.Errorf("got %v want %v", got, tt.valid)
			}
		})
	}
}

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name      string
		user      *User
		valid     bool
		wantPanic bool
	}{
		{
			name:  "valid",
			user:  newTestUser(t),
			valid: true,
		},
		{
			name: "empty name",
			user: func() *User {
				u := newTestUser(t)
				u.Name = ""
				return u
			}(),
			valid: false,
		},
		{
			name: "name too long",
			user: func() *User {
				u := newTestUser(t)
				u.Name = strings.Repeat("a", 501)
				return u
			}(),
			valid: false,
		},
		{
			name: "nil plaintext is allowed",
			user: func() *User {
				u := newTestUser(t)
				u.Password.plaintext = nil
				return u
			}(),
			valid: true,
		},
		{
			name: "missing hash panics",
			user: func() *User {
				u := newTestUser(t)
				u.Password.hash = nil
				return u
			}(),
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					if recover() == nil {
						t.Fatal("expected panic")
					}
				}()
			}
			v := validator.New()
			ValidateUser(v, tt.user)
			if tt.wantPanic {
				return
			}
			if got := v.Valid(); got != tt.valid {
				t.Errorf("got %v want %v", got, tt.valid)
			}
		})
	}
}

func TestUser_IsAnonymous(t *testing.T) {
	if !AnonymousUser.IsAnonymous() {
		t.Error("expected AnonymousUser")
	}
	user := &User{}
	if user.IsAnonymous() {
		t.Error("unexpected anonymous")
	}
}

func BenchmarkPasswordSet(b *testing.B) {
	for b.Loop() {
		var p password
		if err := p.Set("password123"); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPasswordMatches(b *testing.B) {
	var p password

	if err := p.Set("password123"); err != nil {
		b.Fatal(err)
	}

	for b.Loop() {
		ok, err := p.Matches("password123")
		if err != nil || !ok {
			b.Fatal(err)
		}
	}
}
