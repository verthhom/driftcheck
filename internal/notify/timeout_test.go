package notify_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeTimeoutResult(service string) drift.Result {
	return drift.Result{ServiceName: service}
}

func TestTimeoutNotifier_NilInnerPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewTimeoutNotifier(nil, time.Second)
}

func TestTimeoutNotifier_ZeroTimeoutPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero timeout")
		}
	}()
	notify.NewTimeoutNotifier(&stubNotifier{}, 0)
}

func TestTimeoutNotifier_ForwardsSuccessfulCall(t *testing.T) {
	t.Parallel()
	stub := &stubNotifier{}
	tn := notify.NewTimeoutNotifier(stub, 100*time.Millisecond)

	err := tn.Notify(context.Background(), makeTimeoutResult("svc"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", stub.calls)
	}
}

func TestTimeoutNotifier_ForwardsInnerError(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("inner failure")
	stub := &stubNotifier{err: sentinel}
	tn := notify.NewTimeoutNotifier(stub, 100*time.Millisecond)

	err := tn.Notify(context.Background(), makeTimeoutResult("svc"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestTimeoutNotifier_ReturnsErrorOnTimeout(t *testing.T) {
	t.Parallel()
	slow := &slowNotifier{delay: 200 * time.Millisecond}
	tn := notify.NewTimeoutNotifier(slow, 20*time.Millisecond)

	start := time.Now()
	err := tn.Notify(context.Background(), makeTimeoutResult("svc"))
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
	if elapsed > 100*time.Millisecond {
		t.Fatalf("Notify blocked too long: %v", elapsed)
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected DeadlineExceeded in error chain, got %v", err)
	}
}

// slowNotifier blocks for delay before returning.
type slowNotifier struct {
	delay time.Duration
}

func (s *slowNotifier) Notify(ctx context.Context, _ drift.Result) error {
	select {
	case <-time.After(s.delay):
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
