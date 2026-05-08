package notify_test

import (
	"errors"
	"net/smtp"
	"strings"
	"testing"

	"driftcheck/internal/alert"
	"driftcheck/internal/notify"
)

func makeEmailResult(service string, keys []string) alert.Result {
	sev := alert.SeverityFor(keys)
	return alert.Result{
		ServiceName: service,
		Severity:    sev,
		DriftedKeys: keys,
	}
}

func TestEmailNotifier_NoRecipients(t *testing.T) {
	n := notify.NewEmailNotifier(notify.EmailConfig{
		SMTPHost: "localhost",
		SMTPPort: 25,
		From:     "drift@example.com",
	})
	err := n.Notify(makeEmailResult("svc", []string{"KEY"}))
	if err == nil || !strings.Contains(err.Error(), "no recipients") {
		t.Fatalf("expected no-recipients error, got %v", err)
	}
}

func TestEmailNotifier_NoSMTPHost(t *testing.T) {
	n := notify.NewEmailNotifier(notify.EmailConfig{
		To:   []string{"ops@example.com"},
		From: "drift@example.com",
	})
	err := n.Notify(makeEmailResult("svc", []string{"KEY"}))
	if err == nil || !strings.Contains(err.Error(), "SMTP host") {
		t.Fatalf("expected SMTP host error, got %v", err)
	}
}

func TestEmailNotifier_SendsCorrectPayload(t *testing.T) {
	var capturedAddr string
	var capturedMsg []byte

	n := notify.NewEmailNotifier(notify.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "drift@example.com",
		To:       []string{"ops@example.com"},
	})
	n.SetDialFunc(func(addr string, _ smtp.Auth, _ string, _ []string, msg []byte) error {
		capturedAddr = addr
		capturedMsg = msg
		return nil
	})

	err := n.Notify(makeEmailResult("payment-service", []string{"DB_HOST", "API_KEY"}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedAddr != "smtp.example.com:587" {
		t.Errorf("unexpected addr: %s", capturedAddr)
	}
	body := string(capturedMsg)
	if !strings.Contains(body, "payment-service") {
		t.Error("expected service name in message body")
	}
	if !strings.Contains(body, "DB_HOST") {
		t.Error("expected drifted key in message body")
	}
}

func TestEmailNotifier_SMTPError(t *testing.T) {
	n := notify.NewEmailNotifier(notify.EmailConfig{
		SMTPHost: "smtp.example.com",
		SMTPPort: 587,
		From:     "drift@example.com",
		To:       []string{"ops@example.com"},
	})
	n.SetDialFunc(func(_ string, _ smtp.Auth, _ string, _ []string, _ []byte) error {
		return errors.New("connection refused")
	})

	err := n.Notify(makeEmailResult("svc", []string{"KEY"}))
	if err == nil || !strings.Contains(err.Error(), "send failed") {
		t.Fatalf("expected send-failed error, got %v", err)
	}
}
