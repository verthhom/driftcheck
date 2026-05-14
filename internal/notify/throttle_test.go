package notify_test

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

type countingNotifier struct {
	calls atomic.Int32
	err   error
}

func (c *countingNotifier) Notify(_ drift.Result) error {
	c.calls.Add(1)
	return c.err
}

func throttleResult(service string) drift.Result {
	return drift.Result{ServiceName: service}
}

func TestThrottleNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewThrottleNotifier(nil, time.Minute)
}

func TestThrottleNotifier_ZeroWindowAlwaysForwards(t *testing.T) {
	inner := &countingNotifier{}
	tn := notify.NewThrottleNotifier(inner, 0)

	for i := 0; i < 5; i++ {
		if err := tn.Notify(throttleResult("svc")); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if got := inner.calls.Load(); got != 5 {
		t.Fatalf("expected 5 calls, got %d", got)
	}
}

func TestThrottleNotifier_FirstCallAlwaysPasses(t *testing.T) {
	inner := &countingNotifier{}
	tn := notify.NewThrottleNotifier(inner, time.Hour)

	if err := tn.Notify(throttleResult("alpha")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestThrottleNotifier_SuppressesWithinWindow(t *testing.T) {
	inner := &countingNotifier{}
	now := time.Now()
	tn := notify.NewThrottleNotifier(inner, time.Minute)
	tn.SetNow(func() time.Time { return now })

	_ = tn.Notify(throttleResult("beta"))
	err := tn.Notify(throttleResult("beta"))
	if err == nil {
		t.Fatal("expected throttle error on second call")
	}
	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 forwarded call, got %d", got)
	}
}

func TestThrottleNotifier_ForwardsAfterWindowExpires(t *testing.T) {
	inner := &countingNotifier{}
	now := time.Now()
	tn := notify.NewThrottleNotifier(inner, time.Minute)
	tn.SetNow(func() time.Time { return now })

	_ = tn.Notify(throttleResult("gamma"))
	now = now.Add(2 * time.Minute)
	if err := tn.Notify(throttleResult("gamma")); err != nil {
		t.Fatalf("expected forwarded call after window, got: %v", err)
	}
	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 forwarded calls, got %d", got)
	}
}

func TestThrottleNotifier_IndependentPerService(t *testing.T) {
	inner := &countingNotifier{}
	now := time.Now()
	tn := notify.NewThrottleNotifier(inner, time.Hour)
	tn.SetNow(func() time.Time { return now })

	_ = tn.Notify(throttleResult("svc-a"))
	_ = tn.Notify(throttleResult("svc-b"))
	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls (one per service), got %d", got)
	}
}

func TestThrottleNotifier_PropagatesInnerError(t *testing.T) {
	expected := errors.New("inner failure")
	inner := &countingNotifier{err: expected}
	tn := notify.NewThrottleNotifier(inner, 0)

	if err := tn.Notify(throttleResult("svc")); !errors.Is(err, expected) {
		t.Fatalf("expected inner error, got: %v", err)
	}
}
