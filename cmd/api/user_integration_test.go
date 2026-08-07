//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/TTSS0529/greenlight-api/internal/data"
)

func TestRegisterUserHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Users.DB)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	tests := []struct {
		name       string
		body       any
		wantStatus int
	}{
		{
			name: "valid registration",
			body: map[string]any{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "password123",
			},
			wantStatus: http.StatusAccepted,
		},
		{
			name: "duplicate email",
			body: map[string]any{
				"name":     "Alice 2",
				"email":    "alice@example.com",
				"password": "password123",
			},
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			js, err := json.Marshal(tt.body)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := http.Post(
				ts.URL+"/v1/users",
				"application/json",
				bytes.NewBuffer(js),
			)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tt.wantStatus {
				t.Fatalf("got %d want %d", resp.StatusCode, tt.wantStatus)
			}
		})
	}
}

func TestRegisterUserHandler_CreatesUserAndToken(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Users.DB)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	body := map[string]any{
		"name":     "Bob",
		"email":    "bob@example.com",
		"password": "password123",
	}

	js, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		ts.URL+"/v1/users",
		bytes.NewBuffer(js),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted {
		t.Fatalf("got %d want %d", resp.StatusCode, http.StatusAccepted)
	}

	user, err := app.models.Users.GetByEmail("bob@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "Bob" {
		t.Fatalf("got %q want %q", user.Name, "Bob")
	}
	if user.Activated {
		t.Fatal("expected user to be not activated")
	}

	_, err = app.models.Users.GetForToken(data.ScopeActivation, "invalid-token")
	if !errors.Is(err, data.ErrRecordNotFound) {
		t.Fatalf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestActivateUserHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Users.DB)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	user := createTestUser(t, app.models.Users.DB, "Carol", "carol@example.com")
	user.Activated = false
	err := app.models.Users.Update(user)
	if err != nil {
		t.Fatal(err)
	}

	token, err := app.models.Tokens.New(
		user.ID,
		24*time.Hour,
		data.ScopeActivation,
	)
	if err != nil {
		t.Fatal(err)
	}

	body := map[string]any{
		"token": token.Plaintext,
	}

	js, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(
		http.MethodPut,
		ts.URL+"/v1/users/activated",
		bytes.NewBuffer(js),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("got %d want %d", resp.StatusCode, http.StatusOK)
	}

	updated, err := app.models.Users.GetByEmail(user.Email)
	if err != nil {
		t.Fatal(err)
	}
	if !updated.Activated {
		t.Fatal("expected user to be activated")
	}
}

func TestActivateUserHandler_InvalidToken(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Users.DB)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	body := map[string]any{
		"token": "abcdefghijklmnopqrstuvwxyz",
	}

	js, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest(
		http.MethodPut,
		ts.URL+"/v1/users/activated",
		bytes.NewBuffer(js),
	)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("got %d want %d", resp.StatusCode, http.StatusUnprocessableEntity)
	}
}
