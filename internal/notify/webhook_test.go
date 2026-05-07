package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"driftcheck/internal/notify"
)

func TestWebhookNotifier_SendsCorrectPayload(t *testing.T) {
	var received notify.WebhookPayload

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json, got %s", ct)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	p := notify.WebhookPayload{
		Service:   "auth-service",
		Drifted:   true,
		Keys:      []string{"SECRET", "PORT"},
		Severity:  "high",
		Timestamp: time.Now().UTC().Truncate(time.Second),
	}

	if err := n.Notify(context.Background(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received.Service != p.Service {
		t.Errorf("service: want %q, got %q", p.Service, received.Service)
	}
	if !received.Drifted {
		t.Error("expected drifted=true")
	}
	if len(received.Keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(received.Keys))
	}
}

func TestWebhookNotifier_NonSuccessStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL, 5*time.Second)
	err := n.Notify(context.Background(), notify.WebhookPayload{Service: "svc"})
	if err == nil {
		t.Fatal("expected error for 500 response, got nil")
	}
}

func TestWebhookNotifier_EmptyEndpoint(t *testing.T) {
	n := notify.NewWebhookNotifier("", 0)
	err := n.Notify(context.Background(), notify.WebhookPayload{Service: "svc"})
	if err == nil {
		t.Fatal("expected error for empty endpoint, got nil")
	}
}

func TestWebhookNotifier_DefaultTimeout(t *testing.T) {
	// Passing zero timeout should not panic and should still function.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	n := notify.NewWebhookNotifier(srv.URL, 0)
	if err := n.Notify(context.Background(), notify.WebhookPayload{Service: "svc"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
