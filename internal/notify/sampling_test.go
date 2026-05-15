package notify_test

import (
	"errors"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeSamplingResult(hasDrift bool) drift.Result {
	if hasDrift {
		return drift.Result{
			ServiceName: "svc",
			Drifted:     map[string]drift.Delta{"KEY": {Declared: "a", Actual: "b"}},
		}
	}
	return drift.Result{ServiceName: "svc"}
}

func TestSamplingNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewSamplingNotifier(nil, 1.0, 0)
}

func TestSamplingNotifier_InvalidRatePanics(t *testing.T) {
	capture := &captureNotifier{}
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for rate > 1.0")
		}
	}()
	notify.NewSamplingNotifier(capture, 1.5, 0)
}

func TestSamplingNotifier_DriftAlwaysForwarded(t *testing.T) {
	capture := &captureNotifier{}
	// rate=0 means no sampling, but drift must still pass through
	s := notify.NewSamplingNotifier(capture, 0.0, 42)

	result := makeSamplingResult(true)
	if err := s.Notify(result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(capture.results) != 1 {
		t.Fatalf("expected 1 forwarded result, got %d", len(capture.results))
	}
}

func TestSamplingNotifier_NoDriftRateZeroDropsAll(t *testing.T) {
	capture := &captureNotifier{}
	s := notify.NewSamplingNotifier(capture, 0.0, 42)

	for i := 0; i < 20; i++ {
		if err := s.Notify(makeSamplingResult(false)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if len(capture.results) != 0 {
		t.Fatalf("expected 0 forwarded results at rate=0, got %d", len(capture.results))
	}
}

func TestSamplingNotifier_RateOneForwardsAll(t *testing.T) {
	capture := &captureNotifier{}
	s := notify.NewSamplingNotifier(capture, 1.0, 42)

	const n = 10
	for i := 0; i < n; i++ {
		if err := s.Notify(makeSamplingResult(false)); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if len(capture.results) != n {
		t.Fatalf("expected %d forwarded results at rate=1, got %d", n, len(capture.results))
	}
}

func TestSamplingNotifier_PropagatesInnerError(t *testing.T) {
	errInner := errors.New("inner failure")
	failing := &errorNotifier{err: errInner}
	s := notify.NewSamplingNotifier(failing, 1.0, 0)

	if err := s.Notify(makeSamplingResult(true)); !errors.Is(err, errInner) {
		t.Fatalf("expected inner error, got %v", err)
	}
}
