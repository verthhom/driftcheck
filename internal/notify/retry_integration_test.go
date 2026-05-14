package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func TestRetryNotifier_IntegrationWithWebhook(t *testing.T) {
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	webhook := notify.NewWebhookNotifier(server.URL, nil)
	retrier := notify.NewRetryNotifier(webhook, 5, 5*time.Millisecond, nil)

	result := drift.Result{
		ServiceName: "payments",
		Drifted: []drift.KeyDiff{
			{Key: "DB_HOST", Declared: "prod-db", Actual: "staging-db"},
		},
	}

	if err := retrier.Notify(context.Background(), result); err != nil {
		t.Fatalf("expected success after retries, got: %v", err)
	}
	if atomic.LoadInt32(&calls) != 3 {
		t.Fatalf("expected 3 HTTP calls, got %d", atomic.LoadInt32(&calls))
	}
}

func TestRetryNotifier_IntegrationExhausted(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	webhook := notify.NewWebhookNotifier(server.URL, nil)
	retrier := notify.NewRetryNotifier(webhook, 2, 2*time.Millisecond, nil)

	result := drift.Result{ServiceName: "auth"}
	if err := retrier.Notify(context.Background(), result); err == nil {
		t.Fatal("expected error after exhausting retries")
	}
}
