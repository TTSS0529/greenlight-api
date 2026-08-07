package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/data"
	"github.com/TTSS0529/greenlight-api/internal/jsonlog"
)

func newTestApplication(t *testing.T) *application {
	t.Helper()

	return &application{
		logger: jsonlog.New(io.Discard, jsonlog.LevelInfo),
	}
}

func TestRecoverPanic(t *testing.T) {
	app := newTestApplication(t)
	panicHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	})
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	app.recoverPanic(panicHandler).ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("got %d want %d", rr.Code, http.StatusInternalServerError)
	}
	if rr.Header().Get("Connection") != "close" {
		t.Fatal("missing Connection: close")
	}
}

func TestRateLimit_Disabled(t *testing.T) {
	app := newTestApplication(t)
	app.config.limiter.enabled = false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	app.rateLimit(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want %d", rr.Code, http.StatusOK)
	}
}

func TestRateLimit_Exceeded(t *testing.T) {
	app := newTestApplication(t)
	app.config.limiter.enabled = true
	app.config.limiter.rps = 1
	app.config.limiter.burst = 1
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	middleware := app.rateLimit(handler)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "127.0.0.1:1234"
	rr1 := httptest.NewRecorder()
	middleware.ServeHTTP(rr1, req)
	if rr1.Code != http.StatusOK {
		t.Fatalf("first request got %d", rr1.Code)
	}
	rr2 := httptest.NewRecorder()
	middleware.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusTooManyRequests {
		t.Fatalf("second request got %d want %d",
			rr2.Code,
			http.StatusTooManyRequests,
		)
	}
}

func TestAuthenticate_NoAuthorization(t *testing.T) {
	app := newTestApplication(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.contextGetUser(r)
		if !user.IsAnonymous() {
			t.Errorf("expected anonymous user, got %v", user)
		}
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	app.authenticate(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want %d", rr.Code, http.StatusOK)
	}
}

func TestAuthenticate_InvalidHeader(t *testing.T) {
	app := newTestApplication(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Basic abc123")
	rr := httptest.NewRecorder()
	app.authenticate(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticate_MissingBearerToken(t *testing.T) {
	app := newTestApplication(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer")
	rr := httptest.NewRecorder()
	app.authenticate(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticate_InvalidToken(t *testing.T) {
	app := newTestApplication(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next handler should not be called")
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer abc")
	rr := httptest.NewRecorder()
	app.authenticate(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("got %d want %d", rr.Code, http.StatusUnauthorized)
	}
}

func TestAuthenticate_SetsVaryHeader(t *testing.T) {
	app := newTestApplication(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	app.authenticate(handler).ServeHTTP(rr, req)
	if got := rr.Header().Get("Vary"); got != "Authorization" {
		t.Errorf("got Vary header %q want %q", got, "Authorization")
	}
}

func TestRequireAuthenticatedUser(t *testing.T) {
	app := newTestApplication(t)
	tests := []struct {
		name string
		user *data.User
		want int
	}{
		{
			name: "anonymous",
			user: data.AnonymousUser,
			want: http.StatusUnauthorized,
		},
		{
			name: "authenticated",
			user: &data.User{
				ID:        1,
				Activated: true,
			},
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = app.contextSetUser(req, tt.user)
			rr := httptest.NewRecorder()
			app.requireAuthenticatedUser(handler).ServeHTTP(rr, req)
			if rr.Code != tt.want {
				t.Fatalf("got %d want %d", rr.Code, tt.want)
			}
		})
	}
}

func TestRequireActivatedUser(t *testing.T) {
	app := newTestApplication(t)
	tests := []struct {
		name string
		user *data.User
		want int
	}{
		{
			name: "anonymous",
			user: data.AnonymousUser,
			want: http.StatusUnauthorized,
		},
		{
			name: "inactive",
			user: &data.User{
				ID:        1,
				Activated: false,
			},
			want: http.StatusForbidden,
		},
		{
			name: "activated",
			user: &data.User{
				ID:        1,
				Activated: true,
			},
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = app.contextSetUser(req, tt.user)
			rr := httptest.NewRecorder()
			app.requireActivatedUser(handler).ServeHTTP(rr, req)
			if rr.Code != tt.want {
				t.Fatalf("got %d want %d", rr.Code, tt.want)
			}
		})
	}
}

func TestEnableCORS_NoOrigin(t *testing.T) {
	app := newTestApplication(t)
	app.config.cors.trustedOrigins = []string{"http://localhost:3000"}
	nextCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	app.enableCORS(handler).ServeHTTP(rr, req)
	if !nextCalled {
		t.Fatal("next handler not called")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected Access-Control-Allow-Origin header")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want %d", rr.Code, http.StatusOK)
	}
}

func TestEnableCORS_TrustedOrigin(t *testing.T) {
	app := newTestApplication(t)
	app.config.cors.trustedOrigins = []string{"http://localhost:3000"}
	nextCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	rr := httptest.NewRecorder()
	app.enableCORS(handler).ServeHTTP(rr, req)
	if !nextCalled {
		t.Fatal("next handler not called")
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("got %q", got)
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d", rr.Code)
	}
}

func TestEnableCORS_UntrustedOrigin(t *testing.T) {
	app := newTestApplication(t)
	app.config.cors.trustedOrigins = []string{"http://localhost:3000"}
	nextCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Origin", "http://evil.com")
	rr := httptest.NewRecorder()
	app.enableCORS(handler).ServeHTTP(rr, req)
	if !nextCalled {
		t.Fatal("next handler not called")
	}
	if rr.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatal("unexpected Access-Control-Allow-Origin")
	}
}

func TestEnableCORS_Preflight(t *testing.T) {
	app := newTestApplication(t)
	app.config.cors.trustedOrigins = []string{"http://localhost:3000"}
	nextCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
	})
	req := httptest.NewRequest(http.MethodOptions, "/", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	req.Header.Set("Access-Control-Request-Method", "PUT")
	rr := httptest.NewRecorder()
	app.enableCORS(handler).ServeHTTP(rr, req)
	if nextCalled {
		t.Fatal("next handler should not be called")
	}
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:3000" {
		t.Fatalf("got %q", got)
	}
	if rr.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Fatal("missing Allow-Methods")
	}
	if rr.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Fatal("missing Allow-Headers")
	}
}

func TestMetrics(t *testing.T) {
	app := newTestApplication(t)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	app.metrics(handler).ServeHTTP(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d want %d", rr.Code, http.StatusCreated)
	}
}
