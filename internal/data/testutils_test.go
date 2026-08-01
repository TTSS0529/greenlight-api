//go:build integration

package data

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open(
		"postgres",
		os.Getenv("GREENLIGHT_DB_DSN"),
	)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func truncateTables(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`
		TRUNCATE
			tokens,
			users_permissions,
			users,
			movies
		RESTART IDENTITY CASCADE
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func createTestUser(t *testing.T, db *sql.DB, name, email string) *User {
	t.Helper()
	model := UserModel{DB: db}
	user := &User{
		Name:      name,
		Email:     email,
		Activated: true,
	}
	err := user.Password.Set("password123")
	if err != nil {
		t.Fatal(err)
	}
	err = model.Insert(user)
	if err != nil {
		t.Fatal(err)
	}
	return user
}
