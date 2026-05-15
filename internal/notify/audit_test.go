package notify_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeAuditResult(service string, drifted map[string][2]string) drift.Result {
	var diffs []drift.Diff
	for k, v := range drifted {
		diffs = append(diffs, drift.Diff{Key: k, Got: v[0], Want: v[1]})
	}
	return drift.Result{ServiceName: service, Diffs: diffs}
}

type stubNotifier struct {
	err error
}

func (s *stubNotifier) Notify(_ drift.Result) error { return s.err }

func TestAuditNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner notifier")
		}
	}()
	notify.NewAuditNotifier(nil, "test", nil)
}

func TestAuditNotifier_DefaultsToStdout(t *testing.T) {
	// Should not panic when w is nil.
	a := notify.NewAuditNotifier(&stubNotifier{}, "svc", nil)
	if a == nil {
		t.Fatal("expected non-nil AuditNotifier")
	}
}

func TestAuditNotifier_LogsServiceAndNotifierName(t *testing.T) {
	var buf bytes.Buffer
	inner := &stubNotifier{}
	a := notify.NewAuditNotifier(inner, "webhook", &buf)

	result := makeAuditResult("payments", nil)
	if err := a.Notify(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	line := buf.String()
	if !strings.Contains(line, "notifier=webhook") {
		t.Errorf("expected notifier name in log, got: %s", line)
	}
	if !strings.Contains(line, "service=payments") {
		t.Errorf("expected service name in log, got: %s", line)
	}
	if !strings.Contains(line, "status=ok") {
		t.Errorf("expected status=ok in log, got: %s", line)
	}
}

func TestAuditNotifier_LogsErrorStatus(t *testing.T) {
	var buf bytes.Buffer
	inner := &stubNotifier{err: errors.New("connection refused")}
	a := notify.NewAuditNotifier(inner, "slack", &buf)

	result := makeAuditResult("inventory", map[string][2]string{"PORT": {"8080", "9090"}})
	err := a.Notify(result)
	if err == nil {
		t.Fatal("expected error to propagate")
	}

	line := buf.String()
	if !strings.Contains(line, "error: connection refused") {
		t.Errorf("expected error detail in log, got: %s", line)
	}
	if !strings.Contains(line, "drifted=true") {
		t.Errorf("expected drifted=true in log, got: %s", line)
	}
}

func TestAuditNotifier_ForwardsResult(t *testing.T) {
	var buf bytes.Buffer
	inner := &stubNotifier{}
	a := notify.NewAuditNotifier(inner, "email", &buf)

	result := makeAuditResult("auth", nil)
	if err := a.Notify(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(buf.String(), "drifted=false") {
		t.Errorf("expected drifted=false for no-drift result")
	}
}
