//go:build integration

package data

import (
	"database/sql"
	"testing"
	"time"
)

func TestTokenModel_Insert(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := TokenModel{DB: db}
	createTestUser(t, db, "Alice", "alice@example.com")
	token, err := generateToken(
		1,
		time.Hour,
		ScopeActivation,
	)
	if err != nil {
		t.Fatal(err)
	}
	err = model.Insert(token)
	if err != nil {
		t.Fatal(err)
	}
	var count int
	err = db.QueryRow(`
SELECT COUNT(*)
FROM tokens
WHERE user_id=$1
`, 1).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("expected 1 token got %d", count)
	}
}

func TestTokenModel_New(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := TokenModel{DB: db}
	createTestUser(t, db, "Alice", "alice@example.com")
	token, err := model.New(
		1,
		time.Hour,
		ScopeActivation,
	)
	if err != nil {
		t.Fatal(err)
	}
	if token.Plaintext == "" {
		t.Fatal("expected plaintext")
	}
	var count int
	err = db.QueryRow(`
SELECT COUNT(*)
FROM tokens
WHERE hash=$1
`, token.Hash).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatal("token not inserted")
	}
}

func tokenCount(t *testing.T, db *sql.DB, userID int64, scope string) int {
	t.Helper()
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM tokens
		WHERE user_id = $1
		AND scope = $2
	`, userID, scope).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func TestTokenModel_DeleteAllForUser(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := TokenModel{DB: db}
	user1 := createTestUser(t, db, "Alice", "alice@example.com")
	user2 := createTestUser(t, db, "Bob", "Bob@example.com")
	tests := []struct {
		userID int64
		scope  string
	}{
		{user1.ID, ScopeActivation},
		{user1.ID, ScopeAuthentication},
		{user2.ID, ScopeActivation},
	}
	for _, tt := range tests {
		token, err := generateToken(tt.userID, time.Hour, tt.scope)
		if err != nil {
			t.Fatal(err)
		}
		if err := model.Insert(token); err != nil {
			t.Fatal(err)
		}
	}
	err := model.DeleteAllForUser(ScopeActivation, user1.ID)
	if err != nil {
		t.Fatal(err)
	}

	if got := tokenCount(t, db, user1.ID, ScopeActivation); got != 0 {
		t.Fatalf("expected 0 activation tokens for user 1, got %d", got)
	}
	if got := tokenCount(t, db, user1.ID, ScopeAuthentication); got != 1 {
		t.Fatalf("expected 1 authentication token for user 1, got %d", got)
	}
	if got := tokenCount(t, db, user2.ID, ScopeActivation); got != 1 {
		t.Fatalf("expected 1 activation token for user 2, got %d", got)
	}
}
