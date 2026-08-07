//go:build integration

package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/data"
)

func TestCreateMovieHandler(t *testing.T) {
	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name: "valid movie",
			body: `{
				"title":"Inception",
				"year":2010,
				"runtime":"148 mins",
				"genres":["Sci-Fi","Action"]
			}`,
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing title",
			body: `{
				"year":2010,
				"runtime":"148 mins",
				"genres":["Sci-Fi"]
			}`,
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "invalid json",
			body:       `{`,
			wantStatus: http.StatusBadRequest,
		},
	}

	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Movies.DB)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(
				http.MethodPost,
				"/v1/movies",
				strings.NewReader(tt.body),
			)
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()

			app.createMovieHandler(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("want %d got %d", tt.wantStatus, rr.Code)
			}
		})
	}
}

func TestShowMovieHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Movies.DB)

	movie := insertTestMovie(t, app)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "exists",
			id:         strconv.FormatInt(movie.ID, 10),
			wantStatus: http.StatusOK,
		},
		{
			name:       "not found",
			id:         "999999",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "invalid id",
			id:         "abc",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {

		req := httptest.NewRequest(
			http.MethodGet,
			"/v1/movies/"+tt.id,
			nil,
		)
		req = addPathParam(t, req, "id", tt.id)

		rr := httptest.NewRecorder()

		app.showMovieHandler(rr, req)

		if rr.Code != tt.wantStatus {
			t.Fatalf("want %d got %d", tt.wantStatus, rr.Code)
		}
	}
}

func TestUpdateMovieHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Movies.DB)

	tests := []struct {
		name       string
		id         string
		body       string
		wantStatus int
	}{
		{
			name:       "valid update",
			body:       `{"title":"Updated Movie"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       `{"title":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found",
			id:         "999999",
			body:       `{"title":"Updated Movie"}`,
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			movie := insertTestMovie(t, app)

			id := tt.id
			if id == "" {
				id = strconv.FormatInt(movie.ID, 10)
			}

			req := httptest.NewRequest(
				http.MethodPatch,
				"/v1/movies/"+id,
				strings.NewReader(tt.body),
			)

			req.Header.Set("Content-Type", "application/json")
			req = addPathParam(t, req, "id", id)

			rr := httptest.NewRecorder()

			app.updateMovieHandler(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("want %d got %d", tt.wantStatus, rr.Code)
			}

			if rr.Code == http.StatusOK {
				updatedMovie, err := app.models.Movies.Get(movie.ID)

				if err != nil {
					t.Fatal(err)
				}

				if updatedMovie.Title != "Updated Movie" {
					t.Fatalf("want title %q got %q",
						"Updated Movie",
						updatedMovie.Title,
					)
				}
			}
		})
	}
}

func TestDeleteMovieHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Movies.DB)

	tests := []struct {
		name       string
		id         string
		wantStatus int
	}{
		{
			name:       "delete existing movie",
			wantStatus: http.StatusOK,
		},
		{
			name:       "movie not found",
			id:         "999999999",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			movie := insertTestMovie(t, app)

			id := tt.id
			if id == "" {
				id = strconv.FormatInt(movie.ID, 10)
			}

			req := httptest.NewRequest(
				http.MethodDelete,
				"/v1/movies/"+id,
				nil,
			)

			req = addPathParam(t, req, "id", id)

			rr := httptest.NewRecorder()

			app.deleteMovieHandler(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf("want %d got %d", tt.wantStatus, rr.Code)
			}

			if tt.name == "delete existing movie" {
				_, err := app.models.Movies.Get(movie.ID)

				if !errors.Is(err, data.ErrRecordNotFound) {
					t.Fatalf("expected movie to be deleted")
				}
			}
		})
	}
}

func TestListMovieHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Movies.DB)

	insertMulTestMovies(t, app)

	tests := []struct {
		name       string
		url        string
		wantStatus int
		wantMovies int
	}{
		{
			name:       "default list",
			url:        "/v1/movies",
			wantStatus: http.StatusOK,
			wantMovies: 3,
		},
		{
			name:       "filter by title",
			url:        "/v1/movies?title=Matrix",
			wantStatus: http.StatusOK,
			wantMovies: 1,
		},
		{
			name:       "invalid page",
			url:        "/v1/movies?page=abc",
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req := httptest.NewRequest(
				http.MethodGet,
				tt.url,
				nil,
			)

			rr := httptest.NewRecorder()

			app.listMovieHandler(rr, req)

			if rr.Code != tt.wantStatus {
				t.Fatalf(
					"want status %d got %d",
					tt.wantStatus,
					rr.Code,
				)
			}

			if tt.wantMovies > 0 {
				var body struct {
					Movies []data.Movie `json:"movies"`
				}

				err := json.NewDecoder(rr.Body).Decode(&body)
				if err != nil {
					t.Fatal(err)
				}

				if len(body.Movies) != tt.wantMovies {
					t.Fatalf(
						"want %d movies got %d",
						tt.wantMovies,
						len(body.Movies),
					)
				}
			}
		})
	}
}
