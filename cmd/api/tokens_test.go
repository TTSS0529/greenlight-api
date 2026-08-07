package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCreateAuthenticationTokenHandler_MalformedJSON(t *testing.T) {
	app := newTestApplication(t)

	ts := httptest.NewServer(app.routes())
	defer ts.Close()

	req, err := http.NewRequest(
		http.MethodPost,
		ts.URL+"/v1/tokens/authentication",
		bytes.NewBufferString(`{"email":`),
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
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf(
			"want status %d got %d",
			http.StatusBadRequest,
			resp.StatusCode,
		)
	}
}
