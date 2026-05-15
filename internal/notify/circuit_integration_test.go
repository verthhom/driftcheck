package notify_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func TestCircuitBreaker_IntegrationWithWebhook(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	webhook := notify.NewWebhookNotifier(ts.URL)
	cb := notify.NewCircuitBreakerNotifier(webhook, 2, time.Minute)

	result := drift.Result{ServiceName: "svc", HasDrift: true, Drifted: []drift.KeyDiff{{Key: "PORT"}}}

	_ = cb.Notify(result)
	_ = cb.Notify(result)

	if cb.CurrentState() != notify.StateOpen {
		t.Fatalf("expected circuit open after 2 webhook failures")
	}

	err := cb.Notify(result)
	if !errors.Is(err, notify.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen on third call, got %v", err)
	}

	if calls.Load() != 2 {
		t.Fatalf("expected exactly 2 HTTP calls, got %d", calls.Load())
	}
}

func TestCircuitBreaker_IntegrationRecovery(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := calls.Add(1)
		if n <= 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	webhook := notify.NewWebhookNotifier(ts.URL)
	cb := notify.NewCircuitBreakerNotifier(webhook, 1, 5*time.Millisecond)

	result := drift.Result{ServiceName: "svc", HasDrift: true, Drifted: []drift.KeyDiff{{Key: "HOST"}}}

	_ = cb.Notify(result) // fails → opens

	cb.SetOpenedAtDur(10 * time.Millisecond) // simulate cooldown elapsed

	if err := cb.Notify(result); err != nil {
		t.Fatalf("expected recovery call to succeed, got %v", err)
	}
	if cb.CurrentState() != notify.StateClosed {
		t.Fatalf("expected StateClosed after successful recovery, got %v", cb.CurrentState())
	}
}
