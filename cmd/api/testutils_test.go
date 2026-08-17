//go:build integration

package main

import (
	"context"
	"database/sql"
	"net/http"
	"os"
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/data"
	"github.com/julienschmidt/httprouter"
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

func createTestUser(t *testing.T, db *sql.DB, name, email string) *data.User {
	t.Helper()
	model := data.UserModel{DB: db}
	user := &data.User{
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

func insertTestMovie(t *testing.T, app *application) *data.Movie {
	t.Helper()

	movie := &data.Movie{
		Title:   "Test Movie",
		Year:    2024,
		Runtime: data.Runtime(120),
		Genres:  []string{"action", "drama"},
	}

	err := app.models.Movies.Insert(movie)
	if err != nil {
		t.Fatalf("insert movie: %v", err)
	}

	return movie
}

func insertMulTestMovies(t *testing.T, app *application) []*data.Movie {
	t.Helper()

	movies := []*data.Movie{
		{
			Title:   "The Matrix",
			Year:    1999,
			Runtime: 136,
			Genres:  []string{"action", "sci-fi"},
		},
		{
			Title:   "Inception",
			Year:    2010,
			Runtime: 148,
			Genres:  []string{"action", "sci-fi"},
		},
		{
			Title:   "Titanic",
			Year:    1997,
			Runtime: 194,
			Genres:  []string{"drama"},
		},
	}

	for _, movie := range movies {
		err := app.models.Movies.Insert(movie)
		if err != nil {
			t.Fatalf("insert movie: %v", err)
		}
	}

	return movies
}

func addPathParam(t *testing.T, req *http.Request, key, value string) *http.Request {
	t.Helper()

	params := httprouter.Params{
		{
			Key:   key,
			Value: value,
		},
	}
	ctx := context.WithValue(
		req.Context(),
		httprouter.ParamsKey,
		params,
	)

	return req.WithContext(ctx)
}
