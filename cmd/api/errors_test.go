package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func decodeError(t *testing.T, rr *httptest.ResponseRecorder) string {
	t.Helper()

	var body struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	return body.Error
}

func TestErrorResponses(t *testing.T) {
	app := newTestApplication(t)

	tests := []struct {
		name       string
		method     string
		call       func(http.ResponseWriter, *http.Request)
		wantStatus int
		wantBody   string
		wantHeader map[string]string
	}{
		{
			name:   "not found",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.notFoundResponse(w, r)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   "the requested resource could not be found",
		},
		{
			name:   "method not allowed",
			method: http.MethodPost,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.methodNotAllowedResponse(w, r)
			},
			wantStatus: http.StatusMethodNotAllowed,
			wantBody:   "the POST method is not supported for this resource",
		},
		{
			name:   "bad request",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.badRequestResponse(w, r, errors.New("bad json"))
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   "bad json",
		},
		{
			name:   "server error",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.serverErrorResponse(w, r, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   "the server encountered a problem and could not process your request",
		},
		{
			name:   "edit conflict",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.editConflictResponse(w, r)
			},
			wantStatus: http.StatusConflict,
			wantBody:   "unable to update the record due to an edit conflict, please try again",
		},
		{
			name:   "rate limit",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.rateLimitExceededResponse(w, r)
			},
			wantStatus: http.StatusTooManyRequests,
			wantBody:   "rate limit exceeded",
		},
		{
			name:   "invalid credentials",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.invalidCredentialsResponse(w, r)
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid authentication credentials",
		},
		{
			name:   "invalid authentication token",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.invalidAuthenticationTokenResponse(w, r)
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "invalid or missing authentication token",
			wantHeader: map[string]string{
				"WWW-Authenticate": "Bearer",
			},
		},
		{
			name:   "authentication required",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.authenticationRequiredResponse(w, r)
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   "you must be authenticated to access this resource",
		},
		{
			name:   "inactive account",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.inactiveAccountResponse(w, r)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "your user account must be activated to access this resource",
		},
		{
			name:   "not permitted",
			method: http.MethodGet,
			call: func(w http.ResponseWriter, r *http.Request) {
				app.notPermittedResponse(w, r)
			},
			wantStatus: http.StatusForbidden,
			wantBody:   "your user account doesn't have the necessary permissions to access this resource",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, "/", nil)
			rr := httptest.NewRecorder()

			tc.call(rr, req)

			if rr.Code != tc.wantStatus {
				t.Fatalf("got %d want %d", rr.Code, tc.wantStatus)
			}

			got := decodeError(t, rr)
			if got != tc.wantBody {
				t.Fatalf("got %q want %q", got, tc.wantBody)
			}

			for k, v := range tc.wantHeader {
				if rr.Header().Get(k) != v {
					t.Fatalf("%s: got %q want %q", k, rr.Header().Get(k), v)
				}
			}
		})
	}
}

func TestFailedValidationResponse(t *testing.T) {
	app := newTestApplication(t)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()

	want := map[string]string{
		"title": "must be provided",
		"year":  "must be greater than zero",
	}

	app.failedValidationResponse(rr, req, want)

	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("got %d want %d", rr.Code, http.StatusUnprocessableEntity)
	}

	var body struct {
		Error map[string]string `json:"error"`
	}

	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}

	if len(body.Error) != len(want) {
		t.Fatalf("got %d validation errors want %d", len(body.Error), len(want))
	}

	for k, v := range want {
		if body.Error[k] != v {
			t.Fatalf("%s: got %q want %q", k, body.Error[k], v)
		}
	}
}
