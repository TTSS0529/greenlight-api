//go:build integration

package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/data"
	"github.com/TTSS0529/greenlight-api/internal/jsonlog"
)

func newIntegrationApplication(t *testing.T) *application {
	t.Helper()

	db := newTestDB(t)
	return &application{
		logger: jsonlog.New(io.Discard, jsonlog.LevelInfo),
		models: data.NewModels(db),
	}
}

func TestRequirePermission_NoPermission(t *testing.T) {
	app := newIntegrationApplication(t)
	db := app.models.Users.DB
	truncateTables(t, db)
	user := createTestUser(t, db, "Alice", "alice@example.com")
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = app.contextSetUser(req, user)
	rr := httptest.NewRecorder()
	nextCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	app.requirePermission("movies:write", handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("got %d want %d", rr.Code, http.StatusForbidden)
	}
	if nextCalled {
		t.Fatal("next handler should not be called")
	}
}

func TestRequirePermission_WithPermission(t *testing.T) {
	app := newIntegrationApplication(t)
	db := app.models.Users.DB
	truncateTables(t, db)
	user := createTestUser(t, db, "Bob", "bob@example.com")
	err := app.models.Permissions.AddForUser(user.ID, "movies:write")
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = app.contextSetUser(req, user)
	rr := httptest.NewRecorder()
	nextCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	app.requirePermission("movies:write", handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want %d", rr.Code, http.StatusOK)
	}
	if !nextCalled {
		t.Fatal("next handler was not called")
	}
}
