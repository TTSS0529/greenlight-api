package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/TTSS0529/greenlight-api/internal/jsonlog"
	"github.com/TTSS0529/greenlight-api/internal/validator"
	"github.com/julienschmidt/httprouter"
)

func TestReadIDParam(t *testing.T) {
	app := application{}
	tests := []struct {
		name string
		id   string
		want int64
		err  bool
	}{
		{
			name: "valid id",
			id:   "123",
			want: 123,
		},
		{
			name: "invalid id",
			id:   "abc",
			err:  true,
		},
		{
			name: "zero id",
			id:   "0",
			err:  true,
		},
		{
			name: "negative id",
			id:   "-1",
			err:  true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			params := httprouter.Params{
				{
					Key:   "id",
					Value: tt.id,
				},
			}
			req = req.WithContext(
				context.WithValue(
					req.Context(),
					httprouter.ParamsKey,
					params,
				),
			)
			got, err := app.readIDParam(req)
			if tt.err {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if got != tt.want {
				t.Errorf("got %d want %d", got, tt.want)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	app := application{}
	rec := httptest.NewRecorder()
	err := app.writeJSON(
		rec,
		http.StatusCreated,
		envelope{
			"name": "test",
		},
		http.Header{
			"X-Test": []string{"true"},
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if rec.Code != http.StatusCreated {
		t.Errorf("got status %d", rec.Code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Errorf("wrong content type")
	}
	if rec.Header().Get("X-Test") != "true" {
		t.Errorf("custom header missing")
	}
	var body envelope
	err = json.Unmarshal(rec.Body.Bytes(), &body)
	if err != nil {
		t.Fatal(err)
	}
	if body["name"] != "test" {
		t.Errorf("wrong body")
	}
}

func TestReadJSON(t *testing.T) {
	app := application{}
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{
			name: "valid json",
			body: `{"name":"test"}`,
		},
		{
			name:    "empty body",
			body:    ``,
			wantErr: true,
		},
		{
			name:    "malformed json",
			body:    `{"name":`,
			wantErr: true,
		},
		{
			name:    "multiple json",
			body:    `{"name":"a"}{"name":"b"}`,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(
				http.MethodPost,
				"/",
				bytes.NewBufferString(tt.body),
			)
			w := httptest.NewRecorder()
			dst := struct {
				Name string `json:"name"`
			}{}
			err := app.readJSON(w, req, &dst)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestReadJSONUnknownField(t *testing.T) {
	app := application{}
	req := httptest.NewRequest(
		http.MethodPost,
		"/",
		bytes.NewBufferString(
			`{"name":"test","unknown":"field"}`,
		),
	)
	w := httptest.NewRecorder()
	var dst struct {
		Name string `json:"name"`
	}
	err := app.readJSON(w, req, &dst)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestReadString(t *testing.T) {
	app := application{}
	qs := url.Values{}
	got := app.readString(
		qs,
		"name",
		"default",
	)
	if got != "default" {
		t.Errorf("got %q", got)
	}
	qs.Set("name", "bob")
	got = app.readString(
		qs,
		"name",
		"default",
	)
	if got != "bob" {
		t.Errorf("got %q", got)
	}
}

func TestReadCSV(t *testing.T) {
	app := application{}
	qs := url.Values{}
	got := app.readCSV(
		qs,
		"genres",
		[]string{"default"},
	)
	if len(got) != 1 || got[0] != "default" {
		t.Fatal("wrong default")
	}
	qs.Set("genres", "action,comedy")
	got = app.readCSV(
		qs,
		"genres",
		nil,
	)
	if len(got) != 2 {
		t.Fatal("wrong split")
	}
}

func TestReadInt(t *testing.T) {
	app := application{}
	tests := []struct {
		value string
		want  int
		err   bool
	}{
		{
			value: "10",
			want:  10,
		},
		{
			value: "abc",
			err:   true,
		},
	}
	for _, tt := range tests {
		qs := url.Values{}
		qs.Set("page", tt.value)
		v := validator.New()
		got := app.readInt(
			qs,
			"page",
			1,
			v,
		)
		if tt.err {
			if !v.Valid() {
				continue
			}
			t.Fatal("expected validation error")
		}
		if got != tt.want {
			t.Errorf("got %d", got)
		}
	}
}

func TestBackground(t *testing.T) {
	app := application{}
	var done atomic.Bool
	app.background(func() {
		done.Store(true)
	})
	app.wg.Wait()
	if !done.Load() {
		t.Fatal("background did not run")
	}
}

func TestBackgroundRecover(t *testing.T) {
	var buf bytes.Buffer
	app := application{
		logger: jsonlog.New(&buf, jsonlog.LevelInfo),
	}
	app.background(func() {
		panic(errors.New("boom"))
	})
	done := make(chan struct{})
	go func() {
		app.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("background blocked")
	}
}

func newTestApplicationBench(b *testing.B) *application {
	b.Helper()

	return &application{
		logger: jsonlog.New(io.Discard, jsonlog.LevelInfo),
	}
}

func BenchmarkWriteJSON(b *testing.B) {
	app := newTestApplicationBench(b)

	data := envelope{
		"movie": map[string]any{
			"id":     1,
			"title":  "The Matrix",
			"year":   1999,
			"genres": []string{"action", "sci-fi"},
		},
	}

	header := make(http.Header)

	for b.Loop() {
		w := httptest.NewRecorder()

		err := app.writeJSON(
			w,
			http.StatusOK,
			data,
			header,
		)

		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONMarshal(b *testing.B) {
	data := envelope{
		"movie": map[string]any{
			"id":     1,
			"title":  "The Matrix",
			"year":   1999,
			"genres": []string{"action", "sci-fi"},
		},
	}

	for b.Loop() {
		_, err := json.Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkJSONMarshalIndent(b *testing.B) {
	data := envelope{
		"movie": map[string]any{
			"id":     1,
			"title":  "The Matrix",
			"year":   1999,
			"genres": []string{"action", "sci-fi"},
		},
	}

	for b.Loop() {
		_, err := json.MarshalIndent(data, "", "\t")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkReadJSON(b *testing.B) {
	app := newTestApplicationBench(b)

	tests := []struct {
		name string
		body []byte
	}{
		{
			name: "small",
			body: []byte(`{"name":"test"}`),
		},
		{
			name: "large",
			body: []byte(fmt.Sprintf(
				`{"name":"test","description":"%s"}`,
				strings.Repeat("x", 10000),
			)),
		},
	}

	for _, tt := range tests {

		b.Run(tt.name, func(b *testing.B) {

			for b.Loop() {

				req := httptest.NewRequest(
					http.MethodPost,
					"/",
					bytes.NewReader(tt.body),
				)

				w := httptest.NewRecorder()

				var dst map[string]any

				err := app.readJSON(w, req, &dst)

				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
