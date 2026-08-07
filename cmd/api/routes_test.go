package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoutes(t *testing.T) {
	app := newTestApplication(t)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	tests := []struct {
		name       string
		method     string
		url        string
		wantStatus int
	}{
		{
			name:       "GET /v1/healthcheck",
			method:     http.MethodGet,
			url:        "/v1/healthcheck",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST /v1/users",
			method:     http.MethodPost,
			url:        "/v1/users",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "PUT /v1/users/activated",
			method:     http.MethodPut,
			url:        "/v1/users/activated",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "POST /v1/tokens/authentication",
			method:     http.MethodPost,
			url:        "/v1/tokens/authentication",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "GET /debug/vars",
			method:     http.MethodGet,
			url:        "/debug/vars",
			wantStatus: http.StatusOK,
		},
		{
			name:       "unknown route",
			method:     http.MethodGet,
			url:        "/v1/unknown",
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "method not allowed",
			method:     http.MethodPost,
			url:        "/v1/healthcheck",
			wantStatus: http.StatusMethodNotAllowed,
		},
		{
			name:       "GET /v1/movies without authentication",
			method:     http.MethodGet,
			url:        "/v1/movies",
			wantStatus: http.StatusUnauthorized,
		},
	}

	client := ts.Client()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, err := http.NewRequest(
				tt.method,
				ts.URL+tt.url,
				nil,
			)
			if err != nil {
				t.Fatal(err)
			}

			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}

			defer resp.Body.Close()

			// Drain body so httptest connection can be reused
			_, err = io.Copy(io.Discard, resp.Body)
			if err != nil {
				t.Fatal(err)
			}

			if resp.StatusCode != tt.wantStatus {
				t.Errorf(
					"want status %d, got %d",
					tt.wantStatus,
					resp.StatusCode,
				)
			}
		})
	}
}
