package notify_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

type countingNotifier struct {
	calls  int
	failN  int // fail the first N calls
	err    error
}

func (c *countingNotifier) Notify(_ context.Context, _ drift.Result) error {
	c.calls++
	if c.calls <= c.failN {
		return c.err
	}
	return nil
}

func retryResult() drift.Result {
	return drift.Result{ServiceName: "svc"}
}

func TestRetryNotifier_SucceedsFirstAttempt(t *testing.T) {
	inner := &countingNotifier{}
	r := notify.NewRetryNotifier(inner, 3, 0, log.New(os.Stderr, "", 0))
	if err := r.Notify(context.Background(), retryResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestRetryNotifier_RetriesOnFailure(t *testing.T) {
	inner := &countingNotifier{failN: 2, err: errors.New("transient")}
	r := notify.NewRetryNotifier(inner, 3, 0, nil)
	if err := r.Notify(context.Background(), retryResult()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetryNotifier_ReturnsErrorAfterMaxAttempts(t *testing.T) {
	inner := &countingNotifier{failN: 5, err: errors.New("permanent")}
	r := notify.NewRetryNotifier(inner, 3, 0, nil)
	if err := r.Notify(context.Background(), retryResult()); err == nil {
		t.Fatal("expected error, got nil")
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls)
	}
}

func TestRetryNotifier_RespectsContextCancellation(t *testing.T) {
	inner := &countingNotifier{failN: 5, err: errors.New("fail")}
	r := notify.NewRetryNotifier(inner, 5, 50*time.Millisecond, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	if err := r.Notify(ctx, retryResult()); err == nil {
		t.Fatal("expected context error")
	}
}

func TestRetryNotifier_ZeroMaxAttemptsDefaultsToOne(t *testing.T) {
	inner := &countingNotifier{failN: 5, err: errors.New("fail")}
	r := notify.NewRetryNotifier(inner, 0, 0, nil)
	_ = r.Notify(context.Background(), retryResult())
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}
