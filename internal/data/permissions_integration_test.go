//go:build integration

package data

import (
	"slices"
	"testing"
)

func TestPermissionModel_AddForUser(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	user := createTestUser(t, db, "Alice", "alice@example.com")
	model := PermissionModel{DB: db}
	err := model.AddForUser(user.ID, "movies:read")
	if err != nil {
		t.Fatal(err)
	}
	got, err := model.GetAllForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("expected 1 permission, got %d", len(got))
	}
	if got[0] != "movies:read" {
		t.Fatalf("expected movies:read, got %s", got[0])
	}
}

func TestPermissionModel_GetAllForUser(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	user := createTestUser(t, db, "Alice", "alice@example.com")
	model := PermissionModel{DB: db}
	err := model.AddForUser(
		user.ID,
		"movies:read",
		"movies:write",
	)
	if err != nil {
		t.Fatal(err)
	}
	got, err := model.GetAllForUser(user.ID)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(got)
	want := Permissions{
		"movies:read",
		"movies:write",
	}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("want %v got %v", want, got)
	}
}
