package notify_test

import (
	"errors"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func circuitResult(hasDrift bool) drift.Result {
	return drift.Result{
		ServiceName: "svc",
		HasDrift:    hasDrift,
	}
}

func TestCircuitBreaker_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner notifier")
		}
	}()
	notify.NewCircuitBreakerNotifier(nil, 3, time.Second)
}

func TestCircuitBreaker_ClosedForwardsCall(t *testing.T) {
	called := false
	inner := &stubNotifier{fn: func(drift.Result) error { called = true; return nil }}
	cb := notify.NewCircuitBreakerNotifier(inner, 3, time.Minute)

	if err := cb.Notify(circuitResult(false)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected inner notifier to be called")
	}
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	inner := &stubNotifier{fn: func(drift.Result) error { return errors.New("fail") }}
	cb := notify.NewCircuitBreakerNotifier(inner, 3, time.Minute)

	for i := 0; i < 3; i++ {
		_ = cb.Notify(circuitResult(true))
	}

	if cb.CurrentState() != notify.StateOpen {
		t.Fatalf("expected StateOpen, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_RejectsWhenOpen(t *testing.T) {
	inner := &stubNotifier{fn: func(drift.Result) error { return errors.New("fail") }}
	cb := notify.NewCircuitBreakerNotifier(inner, 1, time.Minute)

	_ = cb.Notify(circuitResult(true)) // opens circuit

	err := cb.Notify(circuitResult(true))
	if !errors.Is(err, notify.ErrCircuitOpen) {
		t.Fatalf("expected ErrCircuitOpen, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenAfterCooldown(t *testing.T) {
	inner := &stubNotifier{fn: func(drift.Result) error { return errors.New("fail") }}
	cb := notify.NewCircuitBreakerNotifier(inner, 1, 10*time.Millisecond)

	_ = cb.Notify(circuitResult(true)) // opens circuit
	cb.SetOpenedAtDur(20 * time.Millisecond)

	// Should attempt (half-open) and fail again
	_ = cb.Notify(circuitResult(true))
	if cb.CurrentState() != notify.StateOpen {
		t.Fatalf("expected StateOpen after half-open failure, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_ClosesOnSuccess(t *testing.T) {
	calls := 0
	inner := &stubNotifier{fn: func(drift.Result) error {
		calls++
		if calls < 3 {
			return errors.New("fail")
		}
		return nil
	}}
	cb := notify.NewCircuitBreakerNotifier(inner, 5, 10*time.Millisecond)

	// Accumulate failures but stay closed (threshold=5)
	for i := 0; i < 2; i++ {
		_ = cb.Notify(circuitResult(true))
	}
	// Successful call should reset failures
	if err := cb.Notify(circuitResult(true)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.Failures() != 0 {
		t.Fatalf("expected 0 failures after success, got %d", cb.Failures())
	}
	if cb.CurrentState() != notify.StateClosed {
		t.Fatalf("expected StateClosed, got %v", cb.CurrentState())
	}
}

func TestCircuitBreaker_ZeroThresholdDefaultsToOne(t *testing.T) {
	inner := &stubNotifier{fn: func(drift.Result) error { return errors.New("fail") }}
	cb := notify.NewCircuitBreakerNotifier(inner, 0, time.Minute)

	_ = cb.Notify(circuitResult(true))
	if cb.CurrentState() != notify.StateOpen {
		t.Fatalf("expected StateOpen with threshold=0 (defaults to 1), got %v", cb.CurrentState())
	}
}
