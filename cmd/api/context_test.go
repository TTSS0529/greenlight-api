package main

import (
	"net/http/httptest"
	"testing"

	"github.com/TTSS0529/greenlight-api/internal/data"
)

func TestContextSetGetUser(t *testing.T) {
	app := newTestApplication(t)
	user := &data.User{
		ID:        1,
		Name:      "Alice",
		Email:     "alice@example.com",
		Activated: true,
	}
	req := httptest.NewRequest("", "/", nil)
	req = app.contextSetUser(req, user)
	got := app.contextGetUser(req)
	if got != user {
		t.Errorf("got %+v, want %+v", got, user)
	}
}

func TestContextGetUser_Panic(t *testing.T) {
	app := newTestApplication(t)
	req := httptest.NewRequest("", "/", nil)
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic, got none")
		}
	}()
	app.contextGetUser(req)
}
