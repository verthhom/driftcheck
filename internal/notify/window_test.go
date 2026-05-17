package notify_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"driftcheck/internal/notify"
)

func makeWindowResult(service string, drifted bool) notify.Result {
	keys := map[string]string{}
	if drifted {
		keys["PORT"] = "expected 8080 got 9090"
	}
	return notify.Result{Service: service, Drifted: drifted, DriftedKeys: keys}
}

func TestWindowNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewWindowNotifier(nil, time.Second)
}

func TestWindowNotifier_ZeroWindowPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero window")
		}
	}()
	notify.NewWindowNotifier(&captureNotifier{}, 0)
}

func TestWindowNotifier_EmptyServiceNameErrors(t *testing.T) {
	w := notify.NewWindowNotifier(&captureNotifier{}, 50*time.Millisecond)
	err := w.Notify(context.Background(), notify.Result{})
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestWindowNotifier_AccumulatesLatestPerService(t *testing.T) {
	cap := &captureNotifier{}
	w := notify.NewWindowNotifier(cap, 200*time.Millisecond)
	ctx := context.Background()

	_ = w.Notify(ctx, makeWindowResult("svc-a", false))
	_ = w.Notify(ctx, makeWindowResult("svc-a", true)) // should overwrite
	_ = w.Notify(ctx, makeWindowResult("svc-b", false))

	if got := notify.WindowLatestLen(w); got != 2 {
		t.Fatalf("expected 2 pending, got %d", got)
	}
}

func TestWindowNotifier_FlushForwardsAll(t *testing.T) {
	cap := &captureNotifier{}
	w := notify.NewWindowNotifier(cap, 200*time.Millisecond)
	ctx := context.Background()

	_ = w.Notify(ctx, makeWindowResult("svc-a", true))
	_ = w.Notify(ctx, makeWindowResult("svc-b", false))

	if err := w.Flush(ctx); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}
	if cap.count() != 2 {
		t.Fatalf("expected 2 forwarded, got %d", cap.count())
	}
	if notify.WindowLatestLen(w) != 0 {
		t.Fatal("expected empty buffer after flush")
	}
}

func TestWindowNotifier_FlushReturnsInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	inner := &errorNotifier{err: sentinel}
	w := notify.NewWindowNotifier(inner, 200*time.Millisecond)
	ctx := context.Background()

	_ = w.Notify(ctx, makeWindowResult("svc", true))
	if err := w.Flush(ctx); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestWindowNotifier_TimerAutoFlushes(t *testing.T) {
	var calls atomic.Int32
	inner := notifyFunc(func(_ context.Context, _ notify.Result) error {
		calls.Add(1)
		return nil
	})
	w := notify.NewWindowNotifier(inner, 60*time.Millisecond)
	ctx := context.Background()

	_ = w.Notify(ctx, makeWindowResult("svc", true))

	time.Sleep(150 * time.Millisecond)
	if calls.Load() == 0 {
		t.Fatal("expected auto-flush to have fired")
	}
}
