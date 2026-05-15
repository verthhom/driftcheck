package notify_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

// TestShadowNotifier_IntegrationWithWebhook verifies that the shadow path
// reaches a real HTTP endpoint while the primary result is returned correctly.
func TestShadowNotifier_IntegrationWithWebhook(t *testing.T) {
	var shadowHits int64

	shadowServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt64(&shadowHits, 1)
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("shadow: invalid JSON body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer shadowServer.Close()

	primary := &countNotifier{}
	shadowWebhook := notify.NewWebhookNotifier(shadowServer.URL)

	sn := notify.NewShadowNotifier(primary, shadowWebhook, nil)

	result := drift.Result{
		ServiceName: "payments",
		Drifted: []drift.KeyDiff{
			{Key: "DB_HOST", Declared: "db.prod", Actual: "db.staging"},
		},
	}

	if err := sn.Notify(context.Background(), result); err != nil {
		t.Fatalf("unexpected primary error: %v", err)
	}
	notify.ShadowWait(sn)

	if atomic.LoadInt64(&shadowHits) != 1 {
		t.Errorf("shadow webhook hits = %d, want 1", shadowHits)
	}
	if primary.calls != 1 {
		t.Errorf("primary calls = %d, want 1", primary.calls)
	}
}
