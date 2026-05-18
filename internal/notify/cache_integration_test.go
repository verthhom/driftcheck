package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func TestCacheNotifier_IntegrationWithWebhook(t *testing.T) {
	var calls atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("invalid JSON payload: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	wh := notify.NewWebhookNotifier(ts.URL)
	c := notify.NewCacheNotifier(wh, 50*time.Millisecond)

	r := drift.Result{ServiceName: "integration-svc", Drifted: nil}

	// First call — should reach the webhook.
	if err := c.Notify(r); err != nil {
		t.Fatalf("first notify error: %v", err)
	}
	// Second call same fingerprint — should be suppressed.
	if err := c.Notify(r); err != nil {
		t.Fatalf("second notify error: %v", err)
	}
	if calls.Load() != 1 {
		t.Fatalf("expected 1 webhook call, got %d", calls.Load())
	}

	// Wait for TTL to expire then retry — should reach webhook again.
	time.Sleep(60 * time.Millisecond)
	if err := c.Notify(r); err != nil {
		t.Fatalf("third notify error: %v", err)
	}
	if calls.Load() != 2 {
		t.Fatalf("expected 2 webhook calls after TTL, got %d", calls.Load())
	}
}
