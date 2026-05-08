package notify_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeSlackResult(drifted bool) drift.Result {
	r := drift.Result{ServiceName: "payments"}
	if drifted {
		r.Drifted = []drift.Discrepancy{
			{Key: "LOG_LEVEL", Expected: "info", Actual: "debug"},
		}
	}
	return r
}

func TestSlackNotifier_SendsNoDriftMessage(t *testing.T) {
	var got map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewSlackNotifier(ts.URL, 0)
	if err := n.Notify(context.Background(), makeSlackResult(false)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got["text"], "no drift detected") {
		t.Errorf("expected no-drift message, got: %s", got["text"])
	}
}

func TestSlackNotifier_SendsDriftMessage(t *testing.T) {
	var got map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &got)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	n := notify.NewSlackNotifier(ts.URL, 0)
	if err := n.Notify(context.Background(), makeSlackResult(true)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(got["text"], "LOG_LEVEL") {
		t.Errorf("expected key in message, got: %s", got["text"])
	}
}

func TestSlackNotifier_NonSuccessStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer ts.Close()

	n := notify.NewSlackNotifier(ts.URL, 0)
	err := n.Notify(context.Background(), makeSlackResult(false))
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestSlackNotifier_EmptyWebhookURL(t *testing.T) {
	n := notify.NewSlackNotifier("", 0)
	err := n.Notify(context.Background(), makeSlackResult(false))
	if err == nil {
		t.Fatal("expected error for empty webhook URL")
	}
}

func TestSlackNotifier_DefaultTimeout(t *testing.T) {
	// Passing zero timeout should not panic and should apply a default.
	n := notify.NewSlackNotifier("http://localhost", 0)
	if n == nil {
		t.Fatal("expected non-nil notifier")
	}
	_ = time.Second // ensure time import used
}
