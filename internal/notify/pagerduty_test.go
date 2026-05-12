package notify_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makePDResult(service string, drifted map[string][2]string) drift.Result {
	return drift.Result{
		ServiceName: service,
		Drifted:     drifted,
	}
}

func TestPagerDutyNotifier_EmptyIntegrationKey(t *testing.T) {
	n := notify.NewPagerDutyNotifier("")
	err := n.Notify(makePDResult("svc", map[string][2]string{"K": {"a", "b"}}))
	if err == nil {
		t.Fatal("expected error for empty integration key")
	}
}

func TestPagerDutyNotifier_NoDriftSkipsSend(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := notify.NewPagerDutyNotifier("key123")
	n.(*notify.PagerDutyNotifier).SetEndpoint(ts.URL)

	if err := n.Notify(makePDResult("svc", nil)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no HTTP call when there is no drift")
	}
}

func TestPagerDutyNotifier_SendsCorrectPayload(t *testing.T) {
	var received map[string]interface{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	n := notify.NewPagerDutyNotifier("my-key")
	n.(*notify.PagerDutyNotifier).SetEndpoint(ts.URL)

	err := n.Notify(makePDResult("api-gateway", map[string][2]string{"PORT": {"8080", "9090"}}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if received["routing_key"] != "my-key" {
		t.Errorf("routing_key = %v, want my-key", received["routing_key"])
	}
	if received["event_action"] != "trigger" {
		t.Errorf("event_action = %v, want trigger", received["event_action"])
	}
}

func TestPagerDutyNotifier_NonSuccessStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer ts.Close()

	n := notify.NewPagerDutyNotifier("key")
	n.(*notify.PagerDutyNotifier).SetEndpoint(ts.URL)

	err := n.Notify(makePDResult("svc", map[string][2]string{"X": {"1", "2"}}))
	if err == nil {
		t.Fatal("expected error for non-2xx response")
	}
}
