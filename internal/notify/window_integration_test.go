package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"driftcheck/internal/notify"
)

func TestWindowNotifier_IntegrationWithWebhook(t *testing.T) {
	var received atomic.Int32
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "bad body", http.StatusBadRequest)
			return
		}
		received.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	webhook := notify.NewWebhookNotifier(ts.URL)
	win := notify.NewWindowNotifier(webhook, 80*time.Millisecond)
	ctx := context.Background()

	// Send three updates for the same service — only the latest should be sent.
	for i := 0; i < 3; i++ {
		_ = win.Notify(ctx, notify.Result{Service: "api", Drifted: true,
			DriftedKeys: map[string]string{"PORT": "mismatch"}})
		time.Sleep(5 * time.Millisecond)
	}

	// Also send one for a different service.
	_ = win.Notify(ctx, notify.Result{Service: "worker", Drifted: false})

	if err := win.Flush(ctx); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	// Exactly 2 HTTP calls expected (one per service).
	if got := received.Load(); got != 2 {
		t.Fatalf("expected 2 webhook calls, got %d", got)
	}
}
