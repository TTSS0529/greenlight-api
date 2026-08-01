//go:build integration

package data

import (
	"errors"
	"testing"
	"time"
)

func TestUserModel_Insert(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := UserModel{DB: db}
	user := &User{
		Name:      "Alice",
		Email:     "alice@test.com",
		Activated: true,
	}
	if err := user.Password.Set("password123"); err != nil {
		t.Fatal(err)
	}
	err := model.Insert(user)
	if err != nil {
		t.Fatal(err)
	}
	if user.ID == 0 {
		t.Error("expected id")
	}
	if user.Version != 1 {
		t.Errorf("got %d want 1", user.Version)
	}
}

func TestUserModel_InsertDuplicateEmail(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := UserModel{DB: db}
	u1 := createTestUser(t, db, "Alice", "alice@test.com")
	u2 := &User{
		Name:  "Bob",
		Email: u1.Email,
	}
	if err := u2.Password.Set("password123"); err != nil {
		t.Fatal(err)
	}
	err := model.Insert(u2)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail got %v", err)
	}
}

func TestUserModel_GetByEmail(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	user := createTestUser(t, db, "Alice", "alice@test.com")
	model := UserModel{DB: db}
	got, err := model.GetByEmail(user.Email)
	if err != nil {
		t.Fatal(err)
	}
	if got.Email != user.Email {
		t.Error("email mismatch")
	}
	if got.Name != user.Name {
		t.Error("name mismatch")
	}
}

func TestUserModel_GetByEmail_NotFound(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := UserModel{DB: db}
	_, err := model.GetByEmail("notfound@test.com")
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound got %v", err)
	}
}

func TestUserModel_Update(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := UserModel{DB: db}
	user := createTestUser(t, db, "Alice", "alice@test.com")
	user.Name = "New Name"
	err := model.Update(user)
	if err != nil {
		t.Fatal(err)
	}
	if user.Version != 2 {
		t.Errorf("got %d want 2", user.Version)
	}
}

func TestUserModel_UpdateConflict(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := UserModel{DB: db}
	user := createTestUser(t, db, "Alice", "alice@test.com")
	user.Version = 100
	err := model.Update(user)
	if !errors.Is(err, ErrEditConflict) {
		t.Fatalf("expected ErrEditConflict got %v", err)
	}
}

func TestUserModel_UpdateDuplicateEmail(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := UserModel{DB: db}
	createTestUser(t, db, "Alice", "alice@test.com")
	u2 := createTestUser(t, db, "Bob", "bob@test.com")
	u2.Email = "alice@test.com"
	err := model.Update(u2)
	if !errors.Is(err, ErrDuplicateEmail) {
		t.Fatalf("expected ErrDuplicateEmail got %v", err)
	}
}

func TestUserModel_GetForToken(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	user := createTestUser(t, db, "Alice", "alice@test.com")
	tokenModel := TokenModel{DB: db}
	userModel := UserModel{DB: db}
	token, err := tokenModel.New(
		user.ID,
		24*time.Hour,
		ScopeActivation,
	)
	if err != nil {
		t.Fatal(err)
	}
	got, err := userModel.GetForToken(
		ScopeActivation,
		token.Plaintext,
	)
	if err != nil {
		t.Fatal(err)
	}
	if got.ID != user.ID {
		t.Errorf("got ID %d, want %d", got.ID, user.ID)
	}
	if got.Email != user.Email {
		t.Errorf("got email %q, want %q", got.Email, user.Email)
	}
	if got.Name != user.Name {
		t.Errorf("got name %q, want %q", got.Name, user.Name)
	}
	if got.Activated != user.Activated {
		t.Errorf("activated mismatch")
	}
	if got.Version != user.Version {
		t.Errorf("got version %d, want %d", got.Version, user.Version)
	}
}

func TestUserModel_GetForToken_NotFound(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	userModel := UserModel{DB: db}
	_, err := userModel.GetForToken(
		ScopeActivation,
		"this-token-does-not-exist",
	)
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestUserModel_GetForToken_Expired(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	user := createTestUser(t, db, "Alice", "alice@test.com")
	tokenModel := TokenModel{DB: db}
	userModel := UserModel{DB: db}
	token, err := tokenModel.New(
		user.ID,
		-time.Hour,
		ScopeActivation,
	)
	if err != nil {
		t.Fatal(err)
	}
	if token.Expiry.After(time.Now()) {
		t.Fatal("expected token to be expired")
	}
	_, err = userModel.GetForToken(
		ScopeActivation,
		token.Plaintext,
	)
	if !errors.Is(err, ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}
