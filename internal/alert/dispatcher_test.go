package alert

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/example/driftcheck/internal/drift"
)

type failingNotifier struct{}

func (f *failingNotifier) Notify(_ Alert) error {
	return errors.New("notify failed")
}

func TestDispatcher_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	d := NewDispatcher(NewLogNotifier(&buf))

	result := drift.Result{ServiceName: "svc-a", Diffs: nil}
	if err := d.Dispatch(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "INFO") {
		t.Errorf("expected INFO severity for no-drift, got: %s", out)
	}
}

func TestDispatcher_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	d := NewDispatcher(NewLogNotifier(&buf))

	result := drift.Result{
		ServiceName: "svc-b",
		Diffs: []drift.Diff{
			{Key: "DB_HOST", Expected: "localhost", Actual: "prod-db"},
			{Key: "DB_PORT", Expected: "5432", Actual: "5433"},
		},
	}
	if err := d.Dispatch(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "WARNING") {
		t.Errorf("expected WARNING severity, got: %s", out)
	}
	if !strings.Contains(out, "svc-b") {
		t.Errorf("expected service name in output, got: %s", out)
	}
}

func TestDispatcher_NotifierError(t *testing.T) {
	d := NewDispatcher(&failingNotifier{})
	result := drift.Result{ServiceName: "svc-c"}

	err := d.Dispatch(result)
	if err == nil {
		t.Fatal("expected error from failing notifier")
	}
	if !strings.Contains(err.Error(), "notifier failed") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestDispatcher_MultipleNotifiers(t *testing.T) {
	var buf1, buf2 bytes.Buffer
	d := NewDispatcher(NewLogNotifier(&buf1), NewLogNotifier(&buf2))

	result := drift.Result{ServiceName: "svc-d"}
	if err := d.Dispatch(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if buf1.Len() == 0 || buf2.Len() == 0 {
		t.Error("expected both notifiers to receive the alert")
	}
}
