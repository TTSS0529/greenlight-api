package mailer

import (
	"errors"
	"testing"
	"time"

	"github.com/go-mail/mail/v2"
)

type mockSender struct {
	calls   int
	errs    []error
	lastMsg *mail.Message
}

func (m *mockSender) DialAndSend(msgs ...*mail.Message) error {
	m.calls++
	if len(msgs) > 0 {
		m.lastMsg = msgs[0]
	}
	if len(m.errs) == 0 {
		return nil
	}
	if m.calls <= len(m.errs) {
		return m.errs[m.calls-1]
	}
	return m.errs[len(m.errs)-1]
}

func testMailer(sender *mockSender) Mailer {
	return Mailer{
		dialer: sender,
		sender: "test@example.com",
		sleep:  func(time.Duration) {}, // no-op
		newMessage: func() *mail.Message {
			return mail.NewMessage()
		},
	}
}

var testData = map[string]any{
	"userID":          int64(42),
	"activationToken": "abc123",
}

func TestSendSuccess(t *testing.T) {
	mock := &mockSender{}
	mailer := testMailer(mock)
	err := mailer.Send(
		"user@example.com",
		"user_welcome.tmpl",
		testData,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mock.calls != 1 {
		t.Fatalf("expected 1 call, got %d", mock.calls)
	}
	if mock.lastMsg == nil {
		t.Fatal("expected message to be sent")
	}
}

func TestRetrySuccess(t *testing.T) {
	mock := &mockSender{
		errs: []error{
			errors.New("temporary failure"),
			errors.New("temporary failure"),
			nil,
		},
	}
	mailer := testMailer(mock)
	err := mailer.Send(
		"user@example.com",
		"user_welcome.tmpl",
		testData,
	)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if mock.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", mock.calls)
	}
}

func TestRetryFail(t *testing.T) {
	mock := &mockSender{
		errs: []error{
			errors.New("smtp down"),
			errors.New("smtp down"),
			errors.New("smtp down"),
		},
	}
	mailer := testMailer(mock)
	err := mailer.Send(
		"user@example.com",
		"user_welcome.tmpl",
		testData,
	)
	if err == nil {
		t.Fatal("expected error")
	}
	if mock.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", mock.calls)
	}
}

func TestTemplateNotFound(t *testing.T) {
	mock := &mockSender{}
	mailer := testMailer(mock)
	err := mailer.Send(
		"user@example.com",
		"does_not_exist.tmpl",
		testData,
	)
	if err == nil {
		t.Fatal("expected template error")
	}
	if mock.calls != 0 {
		t.Fatalf("expected no SMTP calls, got %d", mock.calls)
	}
}

func TestTemplateExecuteError(t *testing.T) {
	mock := &mockSender{}
	mailer := testMailer(mock)

	// Missing activationToken
	badData := map[string]any{
		"userID": int64(42),
	}
	err := mailer.Send(
		"user@example.com",
		"user_welcome.tmpl",
		badData,
	)
	if err == nil {
		t.Fatal("expected template execution error")
	}
	if mock.calls != 0 {
		t.Fatalf("expected no SMTP calls, got %d", mock.calls)
	}
}
