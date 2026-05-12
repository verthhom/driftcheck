package notify_test

import (
	"fmt"
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

func makeRLResult(service string) drift.Result {
	return drift.Result{ServiceName: service}
}

func TestRateLimitedNotifier_NilInner(t *testing.T) {
	_, err := notify.NewRateLimitedNotifier(nil, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner notifier")
	}
}

func TestRateLimitedNotifier_ZeroCooldown(t *testing.T) {
	_, err := notify.NewRateLimitedNotifier(&countingNotifier{}, 0)
	if err == nil {
		t.Fatal("expected error for zero cooldown")
	}
}

func TestRateLimitedNotifier_FirstCallPasses(t *testing.T) {
	cn := &countingNotifier{}
	rl, _ := notify.NewRateLimitedNotifier(cn, time.Minute)

	if err := rl.Notify(makeRLResult("svc-a")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := cn.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestRateLimitedNotifier_SecondCallSuppressed(t *testing.T) {
	cn := &countingNotifier{}
	rl, _ := notify.NewRateLimitedNotifier(cn, time.Minute)

	_ = rl.Notify(makeRLResult("svc-a"))
	_ = rl.Notify(makeRLResult("svc-a"))

	if got := cn.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call after suppression, got %d", got)
	}
}

func TestRateLimitedNotifier_DifferentServicesIndependent(t *testing.T) {
	cn := &countingNotifier{}
	rl, _ := notify.NewRateLimitedNotifier(cn, time.Minute)

	_ = rl.Notify(makeRLResult("svc-a"))
	_ = rl.Notify(makeRLResult("svc-b"))

	if got := cn.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls for distinct services, got %d", got)
	}
}

func TestRateLimitedNotifier_ResetAllowsNextCall(t *testing.T) {
	cn := &countingNotifier{}
	rl, _ := notify.NewRateLimitedNotifier(cn, time.Minute)

	_ = rl.Notify(makeRLResult("svc-a"))
	rl.Reset("svc-a")
	_ = rl.Notify(makeRLResult("svc-a"))

	if got := cn.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", got)
	}
}

func TestRateLimitedNotifier_PropagatesInnerError(t *testing.T) {
	cn := &countingNotifier{err: fmt.Errorf("downstream failure")}
	rl, _ := notify.NewRateLimitedNotifier(cn, time.Minute)

	if err := rl.Notify(makeRLResult("svc-a")); err == nil {
		t.Fatal("expected error to be propagated")
	}
}
