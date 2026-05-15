package notify_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func TestRoutingNotifier_IntegrationWithWebhook(t *testing.T) {
	var received int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received++
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	wh := notify.NewWebhookNotifier(ts.URL)
	n := notify.NewRoutingNotifier(nil,
		notify.RouteRule{Match: notify.OnlyDrifted, Notifier: wh},
	)

	clean := drift.Result{ServiceName: "svc"}
	drifted := drift.Result{
		ServiceName: "svc",
		Drifted:     []drift.Delta{{Key: "K", Expected: "x", Actual: "y"}},
	}

	if err := n.Notify(context.Background(), clean); err != nil {
		t.Fatalf("unexpected error on clean result: %v", err)
	}
	if received != 0 {
		t.Fatalf("expected 0 webhook calls for clean result, got %d", received)
	}

	if err := n.Notify(context.Background(), drifted); err != nil {
		t.Fatalf("unexpected error on drifted result: %v", err)
	}
	if received != 1 {
		t.Fatalf("expected 1 webhook call for drifted result, got %d", received)
	}
}
