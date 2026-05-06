package alert

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func TestSeverityFor_NoDrift(t *testing.T) {
	if got := SeverityFor(0); got != SeverityInfo {
		t.Errorf("expected INFO, got %s", got)
	}
}

func TestSeverityFor_FewKeys(t *testing.T) {
	for _, n := range []int{1, 2, 3} {
		if got := SeverityFor(n); got != SeverityWarning {
			t.Errorf("SeverityFor(%d): expected WARNING, got %s", n, got)
		}
	}
}

func TestSeverityFor_ManyKeys(t *testing.T) {
	for _, n := range []int{4, 10, 100} {
		if got := SeverityFor(n); got != SeverityCritical {
			t.Errorf("SeverityFor(%d): expected CRITICAL, got %s", n, got)
		}
	}
}

func TestLogNotifier_WritesAlert(t *testing.T) {
	var buf bytes.Buffer
	n := NewLogNotifier(&buf)

	a := Alert{
		ServiceName: "my-service",
		Severity:    SeverityWarning,
		Message:     "drift detected",
		DriftKeys:   []string{"PORT", "HOST"},
		OccurredAt:  time.Now(),
	}

	if err := n.Notify(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "WARNING") {
		t.Errorf("expected WARNING in output, got: %s", out)
	}
	if !strings.Contains(out, "my-service") {
		t.Errorf("expected service name in output, got: %s", out)
	}
}

func TestLogNotifier_DefaultsToStdout(t *testing.T) {
	n := NewLogNotifier(nil)
	if n.Writer == nil {
		t.Error("expected non-nil writer when nil passed")
	}
}
