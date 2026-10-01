package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthcheckHandler(t *testing.T) {
	app := newTestApplication(t)
	app.config.env = "test"

	req := httptest.NewRequest(http.MethodGet, "/v1/healthcheck", nil)
	rr := httptest.NewRecorder()

	app.healthcheckHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("got %d want %d", rr.Code, http.StatusOK)
	}

	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("got content-type %q want application/json", got)
	}

	var body map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &body)
	if err != nil {
		t.Fatalf("unable to decode response: %v", err)
	}

	if got := body["status"]; got != "available" {
		t.Errorf("got status %v want available", got)
	}

	systemInfo, ok := body["system_info"].(map[string]any)
	if !ok {
		t.Fatalf("system_info is not an object")
	}

	if got := systemInfo["environment"]; got != "test" {
		t.Errorf("got environment %v want test", got)
	}

	if got := systemInfo["version"]; got != version {
		t.Errorf("got version %v want %v", got, version)
	}
}
