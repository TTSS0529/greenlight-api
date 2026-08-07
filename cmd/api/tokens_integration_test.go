//go:build integration

package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/data"
)

func TestCreateAuthenticationTokenHandler(t *testing.T) {
	app := newIntegrationApplication(t)
	truncateTables(t, app.models.Tokens.DB)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()
	user := createTestUser(t, app.models.Tokens.DB, "Test user", "test@example.com")

	tests := []struct {
		name       string
		email      string
		password   string
		wantStatus int
	}{
		{
			name:       "valid credentials",
			email:      user.Email,
			password:   "password123",
			wantStatus: http.StatusCreated,
		},
		{
			name:       "wrong password",
			email:      user.Email,
			password:   "wrongpassword",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "user not found",
			email:      "unknown@example.com",
			password:   "password123",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "invalid email",
			email:      "invalid-email",
			password:   "password123",
			wantStatus: http.StatusUnprocessableEntity,
		},
		{
			name:       "empty password",
			email:      user.Email,
			password:   "",
			wantStatus: http.StatusUnprocessableEntity,
		},
	}

	for _, tt := range tests {

		t.Run(tt.name, func(t *testing.T) {
			input := map[string]string{
				"email":    tt.email,
				"password": tt.password,
			}
			jsonBody, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(
				http.MethodPost,
				ts.URL+"/v1/tokens/authentication",
				bytes.NewReader(jsonBody),
			)
			if err != nil {
				t.Fatal(err)
			}
			req.Header.Set(
				"Content-Type",
				"application/json",
			)
			resp, err := ts.Client().Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != tt.wantStatus {
				t.Errorf(
					"want status %d, got %d",
					tt.wantStatus,
					resp.StatusCode,
				)
			}

			// success case: check token response
			if tt.wantStatus == http.StatusCreated {
				var envelope struct {
					AuthenticationToken data.Token `json:"authentication_token"`
				}
				err = json.NewDecoder(resp.Body).Decode(&envelope)
				if err != nil {
					t.Fatal(err)
				}
				if envelope.AuthenticationToken.Plaintext == "" {
					t.Error("expected token to be returned")
				}
				if envelope.AuthenticationToken.Expiry.IsZero() {
					t.Error("expected expiry to be set")
				}
			}
		})
	}
}
