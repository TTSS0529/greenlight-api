//go:build integration

package data

import (
	"errors"
	"testing"
)

func TestMovieModel_Insert(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := MovieModel{
		DB: db,
	}
	movie := &Movie{
		Title:   "The Matrix",
		Year:    1999,
		Runtime: 136,
		Genres: []string{
			"action",
			"sci-fi",
		},
	}
	err := model.Insert(movie)
	if err != nil {
		t.Fatal(err)
	}
	if movie.ID == 0 {
		t.Errorf("expected ID to be set")
	}
	if movie.Version != 1 {
		t.Errorf(
			"expected version 1, got %d",
			movie.Version,
		)
	}
}

func TestMovieModel_Get(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := MovieModel{DB: db}
	movie := &Movie{
		Title:   "Inception",
		Year:    2010,
		Runtime: 148,
		Genres: []string{
			"action",
			"sci-fi",
		},
	}
	err := model.Insert(movie)
	if err != nil {
		t.Fatal(err)
	}
	got, err := model.Get(movie.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Title != movie.Title {
		t.Errorf(
			"want %s got %s",
			movie.Title,
			got.Title,
		)
	}
	if got.Runtime != movie.Runtime {
		t.Errorf(
			"want %d got %d",
			movie.Runtime,
			got.Runtime,
		)
	}
}

func TestMovieModel_Get_NotFound(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := MovieModel{DB: db}
	_, err := model.Get(999)
	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("expected ErrRecordNotFound")
	}
}

func TestMovieModel_Update(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := MovieModel{DB: db}
	movie := &Movie{
		Title:   "Old title",
		Year:    2000,
		Runtime: 100,
		Genres:  []string{"drama"},
	}
	model.Insert(movie)
	movie.Title = "New title"
	err := model.Update(movie)
	if err != nil {
		t.Fatal(err)
	}
	if movie.Version != 2 {
		t.Errorf(
			"expected version 2 got %d",
			movie.Version,
		)
	}
}

func TestMovieModel_Delete(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := MovieModel{DB: db}
	movie := &Movie{
		Title:   "Avatar",
		Year:    2009,
		Runtime: 160,
		Genres:  []string{"fantasy"},
	}
	model.Insert(movie)
	err := model.Delete(movie.ID)
	if err != nil {
		t.Fatal(err)
	}
	_, err = model.Get(movie.ID)
	if !errors.Is(err, ErrRecordNotFound) {
		t.Errorf("expected not found")
	}
}

func TestMovieModel_GetAll(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()
	truncateTables(t, db)
	model := MovieModel{DB: db}
	movies := []*Movie{
		{
			Title:   "Matrix",
			Year:    1999,
			Runtime: 136,
			Genres:  []string{"action"},
		},
		{
			Title:   "Avatar",
			Year:    2009,
			Runtime: 160,
			Genres:  []string{"fantasy"},
		},
	}
	for _, m := range movies {
		model.Insert(m)
	}
	filters := Filters{
		Page:     1,
		PageSize: 5,
		Sort:     "id",
		SortSafelist: []string{
			"id",
			"title",
			"year",
			"runtime",
			"-id",
			"-title",
			"-year",
			"-runtime",
		},
	}
	got, metadata, err := model.GetAll(
		"",
		[]string{},
		filters,
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Errorf(
			"expected 2 movies got %d",
			len(got),
		)
	}
	if metadata.TotalRecords != 2 {
		t.Errorf(
			"expected total 2 got %d",
			metadata.TotalRecords,
		)
	}
}
