package data

import (
	"bytes"
	"crypto/sha256"
	"testing"
	"time"

	"github.com/TTSS0529/greenlight-api/internal/validator"
)

func TestGenerateToken(t *testing.T) {
	userID := int64(99)
	ttl := 2 * time.Hour
	before := time.Now()
	token, err := generateToken(userID, ttl, ScopeActivation)
	if err != nil {
		t.Fatal(err)
	}
	after := time.Now()
	if token.UserID != userID {
		t.Errorf("expected userID %d, got %d", userID, token.UserID)
	}
	if token.Scope != ScopeActivation {
		t.Errorf("expected scope %q, got %q", ScopeActivation, token.Scope)
	}
	if len(token.Plaintext) != 26 {
		t.Errorf("expected plaintext length 26, got %d", len(token.Plaintext))
	}
	expectedHash := sha256.Sum256([]byte(token.Plaintext))
	if !bytes.Equal(expectedHash[:], token.Hash) {
		t.Error("hash mismatch")
	}
	if token.Expiry.Before(before.Add(ttl)) ||
		token.Expiry.After(after.Add(ttl)) {
		t.Error("expiry outside expected range")
	}
}

func TestGenerateToken_Unique(t *testing.T) {
	token1, err := generateToken(1, time.Hour, ScopeActivation)
	if err != nil {
		t.Fatal(err)
	}
	token2, err := generateToken(1, time.Hour, ScopeActivation)
	if err != nil {
		t.Fatal(err)
	}
	if token1.Plaintext == token2.Plaintext {
		t.Fatal("expected different plaintext")
	}
	if bytes.Equal(token1.Hash, token2.Hash) {
		t.Fatal("expected different hash")
	}
}

func TestValidateTokenPlaintext(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		wantValid bool
	}{
		{
			"valid",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ",
			true,
		},
		{
			"empty",
			"",
			false,
		},
		{
			"too short",
			"abc",
			false,
		},
		{
			"too long",
			"ABCDEFGHIJKLMNOPQRSTUVWXYZ123",
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := validator.New()
			ValidateTokenPlaintext(v, tt.token)
			if v.Valid() != tt.wantValid {
				t.Fatalf("expected valid=%v got %v",
					tt.wantValid,
					v.Valid())
			}
		})
	}
}

func BenchmarkGenerateToken(b *testing.B) {
	userID := int64(1)
	scope := ScopeActivation

	b.ResetTimer()

	for b.Loop() {
		_, err := generateToken(userID, time.Hour, scope)
		if err != nil {
			b.Fatal(err)
		}
	}
}
