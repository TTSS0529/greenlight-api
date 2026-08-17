//go:build integration

package data

import (
	"database/sql"
	"testing"
	"time"
)

func favoriteCount(t *testing.T, db *sql.DB, userID int64, movieID int64) int {
	t.Helper()
	var count int
	err := db.QueryRow(`
		SELECT COUNT(*)
		FROM favorite_movies
		WHERE user_id = $1 AND movie_id = $2
	`, userID, movieID).Scan(&count)
	if err != nil {
		t.Fatal(err)
	}
	return count
}

func TestFavoriteModel_Add(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)

	favorites := FavoriteModel{DB: db}

	user := createTestUser(t, db, "tester", "test@example.com")
	movie := createTestMovie(t, db, "test movie")

	// Add favorite.
	err := favorites.Add(user.ID, movie.ID)
	if err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	// Verify that the favorite was actually inserted.
	if count := favoriteCount(t, db, user.ID, movie.ID); count != 1 {
		t.Errorf("expected 1 favorite, got %d", count)
	}
}

func TestFavoriteModel_AddDuplicate(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)

	favorites := FavoriteModel{DB: db}

	user := createTestUser(t, db, "tester", "test@example.com")
	movie := createTestMovie(t, db, "test movie")

	// First add.
	if err := favorites.Add(user.ID, movie.ID); err != nil {
		t.Fatalf("first Add() error = %v", err)
	}

	// Second add should not return an error.
	if err := favorites.Add(user.ID, movie.ID); err != nil {
		t.Fatalf("second Add() error = %v", err)
	}

	if count := favoriteCount(t, db, user.ID, movie.ID); count != 1 {
		t.Errorf("expected 1 favorite, got %d", count)
	}
}

func TestFavoriteModel_Remove(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)

	favorites := FavoriteModel{DB: db}

	user := createTestUser(t, db, "tester", "test@example.com")
	movie := createTestMovie(t, db, "test movie")

	if err := favorites.Add(user.ID, movie.ID); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	if err := favorites.Remove(user.ID, movie.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}

	if count := favoriteCount(t, db, user.ID, movie.ID); count != 0 {
		t.Errorf("expected 0 favorites, got %d", count)
	}
}

func TestFavoriteModel_RemoveNotFound(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)

	favorites := FavoriteModel{DB: db}

	user := createTestUser(t, db, "tester", "test@example.com")
	movie := createTestMovie(t, db, "test movie")

	// Removing a non-existing favorite should not be an error.
	if err := favorites.Remove(user.ID, movie.ID); err != nil {
		t.Fatalf("Remove() error = %v", err)
	}
}

func TestFavoriteModel_GetAll(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)

	favorites := FavoriteModel{DB: db}

	user := createTestUser(t, db, "tester", "test@example.com")
	movie1 := createTestMovie(t, db, "test movie1")
	movie2 := createTestMovie(t, db, "test movie2")

	if err := favorites.Add(user.ID, movie1.ID); err != nil {
		t.Fatalf("Add() movie1 error = %v", err)
	}

	if err := favorites.Add(user.ID, movie2.ID); err != nil {
		t.Fatalf("Add() movie2 error = %v", err)
	}

	// Make movie1 the older favorite.
	_, err := db.Exec(`
		UPDATE favorite_movies
		SET created_at = $1
		WHERE user_id = $2 AND movie_id = $3
	`,
		time.Now().Add(-time.Hour),
		user.ID,
		movie1.ID,
	)
	if err != nil {
		t.Fatalf("failed to update created_at: %v", err)
	}

	got, err := favorites.GetAll(user.ID)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 favorites, got %d", len(got))
	}

	if got[0].ID != movie2.ID {
		t.Errorf("expected first movie ID %d, got %d", movie2.ID, got[0].ID)
	}

	if got[1].ID != movie1.ID {
		t.Errorf("expected second movie ID %d, got %d", movie1.ID, got[1].ID)
	}
}

func TestFavoriteModel_GetAllUserIsolation(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)

	favorites := FavoriteModel{DB: db}

	user1 := createTestUser(t, db, "tester1", "test1@example.com")
	user2 := createTestUser(t, db, "tester2", "test2@example.com")
	movie1 := createTestMovie(t, db, "test movie1")
	movie2 := createTestMovie(t, db, "test movie2")

	if err := favorites.Add(user1.ID, movie1.ID); err != nil {
		t.Fatalf("Add() user1 error = %v", err)
	}

	if err := favorites.Add(user2.ID, movie2.ID); err != nil {
		t.Fatalf("Add() user2 error = %v", err)
	}

	got, err := favorites.GetAll(user1.ID)
	if err != nil {
		t.Fatalf("GetAll() error = %v", err)
	}

	if len(got) != 1 {
		t.Fatalf("expected 1 favorite, got %d", len(got))
	}

	if got[0].ID != movie1.ID {
		t.Errorf("expected movie ID %d, got %d", movie1.ID, got[0].ID)
	}
}
