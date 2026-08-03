package jsonlog

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"
)

type logLine struct {
	Level      string            `json:"level"`
	Time       string            `json:"time"`
	Message    string            `json:"message"`
	Properties map[string]string `json:"properties"`
	Trace      string            `json:"trace"`
}

func TestLevel_String(t *testing.T) {
	tests := []struct {
		level Level
		want  string
	}{
		{LevelInfo, "INFO"},
		{LevelError, "ERROR"},
		{LevelFatal, "FATAL"},
		{Level(100), ""},
	}

	for _, tt := range tests {
		if got := tt.level.String(); got != tt.want {
			t.Errorf("got %q want %q", got, tt.want)
		}
	}
}

func TestNew(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := New(buf, LevelError)
	if logger.out != buf {
		t.Fatal("unexpected writer")
	}
	if logger.minLevel != LevelError {
		t.Fatal("unexpected min level")
	}
}

func TestLogger_PrintInfo(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)
	props := map[string]string{
		"user": "alice",
	}
	logger.PrintInfo("hello", props)
	var got logLine
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Level != "INFO" {
		t.Errorf("got %q", got.Level)
	}
	if got.Message != "hello" {
		t.Errorf("got %q", got.Message)
	}
	if got.Properties["user"] != "alice" {
		t.Error("properties not written")
	}
	if got.Trace != "" {
		t.Error("info log should not contain stack trace")
	}
}

func TestLogger_PrintError(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)
	logger.PrintError(errors.New("database failed"), nil)
	var got logLine
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Level != "ERROR" {
		t.Errorf("got %q", got.Level)
	}
	if got.Message != "database failed" {
		t.Errorf("got %q", got.Message)
	}
	if got.Trace == "" {
		t.Error("expected stack trace")
	}
}

func TestLogger_Write(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)
	n, err := logger.Write([]byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	if n == 0 {
		t.Fatal("expected bytes written")
	}
	var got logLine
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Level != "ERROR" {
		t.Errorf("got %q", got.Level)
	}
	if got.Message != "hello" {
		t.Errorf("got %q", got.Message)
	}
	if got.Trace == "" {
		t.Error("expected trace")
	}
}

func TestLogger_MinLevel(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelError)
	logger.PrintInfo("ignored", nil)
	if buf.Len() != 0 {
		t.Fatal("expected no output")
	}
}

func TestLogger_EmptyProperties(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)
	logger.PrintInfo("hello", nil)
	output := buf.String()
	if strings.Contains(output, "properties") {
		t.Fatal("properties should be omitted")
	}
}
