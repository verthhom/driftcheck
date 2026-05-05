package snapshot_test

import (
	"testing"
	"time"

	"driftcheck/internal/snapshot"
)

func TestNew_CopiesConfig(t *testing.T) {
	orig := map[string]string{"port": "8080", "env": "prod"}
	snap := snapshot.New("svc", orig)

	// Mutating the original map must not affect the snapshot.
	orig["port"] = "9999"

	if v, _ := snap.Get("port"); v != "8080" {
		t.Errorf("expected port 8080, got %s", v)
	}
}

func TestNew_SetsServiceName(t *testing.T) {
	snap := snapshot.New("my-service", nil)
	if snap.ServiceName != "my-service" {
		t.Errorf("expected service name 'my-service', got %q", snap.ServiceName)
	}
}

func TestNew_CapturedAtIsRecent(t *testing.T) {
	before := time.Now().UTC()
	snap := snapshot.New("svc", nil)
	after := time.Now().UTC()

	if snap.CapturedAt.Before(before) || snap.CapturedAt.After(after) {
		t.Errorf("CapturedAt %v is outside expected range [%v, %v]",
			snap.CapturedAt, before, after)
	}
}

func TestGet_MissingKey(t *testing.T) {
	snap := snapshot.New("svc", map[string]string{"a": "1"})
	_, ok := snap.Get("missing")
	if ok {
		t.Error("expected ok=false for missing key")
	}
}
