package drift_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"driftcheck/internal/drift"
)

func TestReporter_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	r := drift.NewReporter(&buf)

	result := drift.Result{
		ServiceName: "api",
		DriftedKeys: nil,
		CheckedAt:   time.Now(),
	}

	r.Report(result)

	got := buf.String()
	if !strings.Contains(got, "[OK]") {
		t.Errorf("expected [OK] prefix, got: %s", got)
	}
	if !strings.Contains(got, "api") {
		t.Errorf("expected service name in output, got: %s", got)
	}
	if !strings.Contains(got, "no drift detected") {
		t.Errorf("expected 'no drift detected' message, got: %s", got)
	}
}

func TestReporter_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	r := drift.NewReporter(&buf)

	result := drift.Result{
		ServiceName: "worker",
		DriftedKeys: []drift.DriftedKey{
			{Key: "LOG_LEVEL", Expected: "info", Actual: "debug", Reason: "value mismatch"},
			{Key: "TIMEOUT", Expected: "30s", Actual: "", Reason: "missing in snapshot"},
		},
		CheckedAt: time.Now(),
	}

	r.Report(result)

	got := buf.String()
	if !strings.Contains(got, "[DRIFT]") {
		t.Errorf("expected [DRIFT] prefix, got: %s", got)
	}
	if !strings.Contains(got, "2 key(s) drifted") {
		t.Errorf("expected drift count in output, got: %s", got)
	}
	if !strings.Contains(got, "LOG_LEVEL") {
		t.Errorf("expected LOG_LEVEL in output, got: %s", got)
	}
	if !strings.Contains(got, "TIMEOUT") {
		t.Errorf("expected TIMEOUT in output, got: %s", got)
	}
}

func TestReporter_DefaultsToStdout(t *testing.T) {
	// Should not panic when nil writer is passed.
	r := drift.NewReporter(nil)
	if r == nil {
		t.Fatal("expected non-nil reporter")
	}
}

func TestResult_HasDrift(t *testing.T) {
	noDrift := drift.Result{}
	if noDrift.HasDrift() {
		t.Error("expected HasDrift to be false for empty result")
	}

	withDrift := drift.Result{
		DriftedKeys: []drift.DriftedKey{{Key: "X"}},
	}
	if !withDrift.HasDrift() {
		t.Error("expected HasDrift to be true when keys are present")
	}
}
