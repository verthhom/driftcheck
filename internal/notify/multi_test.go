package notify_test

import (
	"errors"
	"testing"

	"driftcheck/internal/alert"
	"driftcheck/internal/notify"
)

type stubNotifier struct {
	called int
	err    error
}

func (s *stubNotifier) Notify(_ alert.Result) error {
	s.called++
	return s.err
}

func driftResult() alert.Result {
	return alert.Result{
		ServiceName: "test-svc",
		Severity:    alert.SeverityFor([]string{"KEY"}),
		DriftedKeys: []string{"KEY"},
	}
}

func TestMultiNotifier_CallsAll(t *testing.T) {
	a, b := &stubNotifier{}, &stubNotifier{}
	m := notify.NewMultiNotifier(false, a, b)
	if err := m.Notify(driftResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.called != 1 || b.called != 1 {
		t.Errorf("expected each notifier called once, got a=%d b=%d", a.called, b.called)
	}
}

func TestMultiNotifier_CollectsErrors(t *testing.T) {
	a := &stubNotifier{err: errors.New("fail-a")}
	b := &stubNotifier{err: errors.New("fail-b")}
	m := notify.NewMultiNotifier(false, a, b)
	err := m.Notify(driftResult())
	if err == nil {
		t.Fatal("expected combined error")
	}
	if a.called != 1 || b.called != 1 {
		t.Error("expected both notifiers called even on error")
	}
}

func TestMultiNotifier_StopOnErr(t *testing.T) {
	a := &stubNotifier{err: errors.New("fail-a")}
	b := &stubNotifier{}
	m := notify.NewMultiNotifier(true, a, b)
	err := m.Notify(driftResult())
	if err == nil {
		t.Fatal("expected error")
	}
	if b.called != 0 {
		t.Error("expected second notifier not called after stop-on-err")
	}
}

func TestMultiNotifier_NoNotifiers(t *testing.T) {
	m := notify.NewMultiNotifier(false)
	if err := m.Notify(driftResult()); err != nil {
		t.Fatalf("unexpected error with empty notifier list: %v", err)
	}
}
