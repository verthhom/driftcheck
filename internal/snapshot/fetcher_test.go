package snapshot_test

import (
	"testing"

	"driftcheck/internal/snapshot"
)

func TestEnvFetcher_EmptyServiceName(t *testing.T) {
	f := snapshot.NewEnvFetcher()
	_, err := f.Fetch("")
	if err == nil {
		t.Fatal("expected error for empty service name, got nil")
	}
}

func TestEnvFetcher_ReadsMatchingVars(t *testing.T) {
	t.Setenv("MYSVC_PORT", "3000")
	t.Setenv("MYSVC_LOG_LEVEL", "debug")
	t.Setenv("OTHER_PORT", "9000") // should be ignored

	f := snapshot.NewEnvFetcher()
	snap, err := f.Fetch("mysvc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if snap.ServiceName != "mysvc" {
		t.Errorf("expected service name 'mysvc', got %q", snap.ServiceName)
	}

	if v, ok := snap.Get("port"); !ok || v != "3000" {
		t.Errorf("expected port=3000, got %q (ok=%v)", v, ok)
	}

	if v, ok := snap.Get("log_level"); !ok || v != "debug" {
		t.Errorf("expected log_level=debug, got %q (ok=%v)", v, ok)
	}

	if _, ok := snap.Get("other_port"); ok {
		t.Error("other_port should not be present in snapshot")
	}
}

func TestEnvFetcher_NoMatchingVars(t *testing.T) {
	f := snapshot.NewEnvFetcher()
	snap, err := f.Fetch("nonexistentsvc")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(snap.Config) != 0 {
		t.Errorf("expected empty config, got %v", snap.Config)
	}
}
