package notify_test

import (
	"bytes"
	"strings"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeLogResult(service string, diffs []drift.Diff) drift.Result {
	return drift.Result{ServiceName: service, Diffs: diffs}
}

func TestLogNotifier_NilInnerPanics(t *testing.T) {
	// LogNotifier has no inner dependency; just verify construction succeeds.
	n := notify.NewLogNotifier()
	if n == nil {
		t.Fatal("expected non-nil LogNotifier")
	}
}

func TestLogNotifier_DefaultsToStdout(t *testing.T) {
	// Smoke test: calling Notify with no writer option must not panic.
	n := notify.NewLogNotifier()
	err := n.Notify([]drift.Result{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogNotifier_WritesOkStatus(t *testing.T) {
	var buf bytes.Buffer
	n := notify.NewLogNotifier(notify.WithLogWriter(&buf))

	err := n.Notify([]drift.Result{makeLogResult("svc-a", nil)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "service=svc-a") {
		t.Errorf("expected service name in output, got: %s", got)
	}
	if !strings.Contains(got, "status=ok") {
		t.Errorf("expected status=ok in output, got: %s", got)
	}
}

func TestLogNotifier_WritesDriftStatus(t *testing.T) {
	var buf bytes.Buffer
	n := notify.NewLogNotifier(notify.WithLogWriter(&buf))

	diffs := []drift.Diff{{Key: "PORT", Declared: "8080", Actual: "9090"}}
	err := n.Notify([]drift.Result{makeLogResult("svc-b", diffs)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "status=drift") {
		t.Errorf("expected status=drift in output, got: %s", got)
	}
	if !strings.Contains(got, "drifted_keys=1") {
		t.Errorf("expected drifted_keys=1 in output, got: %s", got)
	}
}

func TestLogNotifier_PrefixAppearsInOutput(t *testing.T) {
	var buf bytes.Buffer
	n := notify.NewLogNotifier(
		notify.WithLogWriter(&buf),
		notify.WithLogPrefix("PROD"),
	)

	_ = n.Notify([]drift.Result{makeLogResult("svc-c", nil)})

	if !strings.Contains(buf.String(), "[PROD]") {
		t.Errorf("expected prefix [PROD] in output, got: %s", buf.String())
	}
}

func TestLogNotifier_MultipleResults(t *testing.T) {
	var buf bytes.Buffer
	n := notify.NewLogNotifier(notify.WithLogWriter(&buf))

	results := []drift.Result{
		makeLogResult("svc-x", nil),
		makeLogResult("svc-y", []drift.Diff{{Key: "DB_URL"}}),
	}
	if err := n.Notify(results); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 log lines, got %d: %s", len(lines), buf.String())
	}
}
