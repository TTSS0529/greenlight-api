//go:build integration

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/data"
)

func TestAddFavoriteHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	db := app.models.Movies.DB
	truncateTables(t, db)

	user := createTestUser(t, db, "tester", "test@example.com")
	movie := insertTestMovie(t, app)

	t.Run("valid", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			fmt.Sprintf("/v1/movies/%d/favorite", movie.ID),
			nil,
		)
		id := strconv.FormatInt(movie.ID, 10)
		req = app.contextSetUser(req, user)
		req = addPathParam(t, req, "id", id)
		rr := httptest.NewRecorder()

		app.addFavoriteHandler(rr, req)

		if rr.Code != http.StatusCreated {
			t.Errorf("want status %d, got %d", http.StatusCreated, rr.Code)
		}

		var response struct {
			Message string `json:"message"`
		}

		err := json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Fatal(err)
		}

		if response.Message != "movie added to favorites" {
			t.Errorf("want message %q, got %q",
				"movie added to favorites",
				response.Message,
			)
		}

		// Verify that the favorite was actually created in the database.
		favorites, err := app.models.Favorites.GetAll(user.ID)
		if err != nil {
			t.Fatal(err)
		}

		if len(favorites) != 1 {
			t.Errorf("want 1 favorite, got %d", len(favorites))
		}

		if favorites[0].ID != movie.ID {
			t.Errorf("want movie ID %d, got %d",
				movie.ID,
				favorites[0].ID,
			)
		}
	})

	t.Run("invalid movie ID", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/v1/movies/not-an-integer/favorite",
			nil,
		)

		id := "not-an-integer"
		req = app.contextSetUser(req, user)
		req = addPathParam(t, req, "id", id)

		rr := httptest.NewRecorder()

		app.addFavoriteHandler(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("want status %d, got %d", http.StatusNotFound, rr.Code)
		}
	})

	t.Run("movie not found", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodPost,
			"/v1/movies/999999/favorite",
			nil,
		)

		id := strconv.FormatInt(999999, 10)
		req = app.contextSetUser(req, user)
		req = addPathParam(t, req, "id", id)

		rr := httptest.NewRecorder()

		app.addFavoriteHandler(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("want status %d, got %d", http.StatusNotFound, rr.Code)
		}
	})
}

func TestRemoveFavoriteHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	db := app.models.Movies.DB
	truncateTables(t, db)

	user := createTestUser(t, db, "tester", "test@example.com")
	movie := insertTestMovie(t, app)

	// First create the favorite that we are going to remove.
	err := app.models.Favorites.Add(user.ID, movie.ID)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("valid", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodDelete,
			fmt.Sprintf("/v1/movies/%d/favorite", movie.ID),
			nil,
		)

		id := strconv.FormatInt(movie.ID, 10)
		req = app.contextSetUser(req, user)
		req = addPathParam(t, req, "id", id)

		rr := httptest.NewRecorder()

		app.removeFavoriteHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d, got %d", http.StatusOK, rr.Code)
		}

		var response struct {
			Message string `json:"message"`
		}

		err := json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Fatal(err)
		}

		if response.Message != "movie removed from favorites" {
			t.Errorf("want message %q, got %q",
				"movie removed from favorites",
				response.Message,
			)
		}

		// Verify that the favorite was actually removed.
		favorites, err := app.models.Favorites.GetAll(user.ID)
		if err != nil {
			t.Fatal(err)
		}

		if len(favorites) != 0 {
			t.Errorf("want 0 favorites, got %d", len(favorites))
		}
	})

	t.Run("invalid movie ID", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodDelete,
			"/v1/movies/not-an-integer/favorite",
			nil,
		)

		id := "not-an-interger"
		req = app.contextSetUser(req, user)
		req = addPathParam(t, req, "id", id)

		rr := httptest.NewRecorder()

		app.removeFavoriteHandler(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("want status %d, got %d", http.StatusNotFound, rr.Code)
		}
	})
}

func TestListFavoriteHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	db := app.models.Movies.DB
	truncateTables(t, db)

	user := createTestUser(t, db, "tester", "test@example.com")
	movies := insertMulTestMovies(t, app)
	for _, movie := range movies {
		err := app.models.Favorites.Add(user.ID, movie.ID)
		if err != nil {
			t.Fatal(err)
		}
	}

	t.Run("valid", func(t *testing.T) {
		req := httptest.NewRequest(
			http.MethodGet,
			"/v1/favorites",
			nil,
		)

		req = app.contextSetUser(req, user)

		rr := httptest.NewRecorder()

		app.listFavoriteHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d, got %d", http.StatusOK, rr.Code)
		}

		var response struct {
			Favorites []data.Favorite `json:"favorites"`
		}

		err := json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Fatal(err)
		}

		if len(response.Favorites) != 3 {
			t.Errorf("want 3 favorites, got %d", len(response.Favorites))
		}
	})

	t.Run("no favorites", func(t *testing.T) {
		otherUser := createTestUser(t, db, "other", "other@example.com")

		req := httptest.NewRequest(
			http.MethodGet,
			"/v1/favorites",
			nil,
		)

		req = app.contextSetUser(req, otherUser)

		rr := httptest.NewRecorder()

		app.listFavoriteHandler(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("want status %d, got %d", http.StatusOK, rr.Code)
		}

		var response struct {
			Favorites []data.Favorite `json:"favorites"`
		}

		err := json.NewDecoder(rr.Body).Decode(&response)
		if err != nil {
			t.Fatal(err)
		}

		if len(response.Favorites) != 0 {
			t.Errorf("want 0 favorites, got %d",
				len(response.Favorites))
		}
	})
}
