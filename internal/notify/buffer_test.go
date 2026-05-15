package notify_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"driftcheck/internal/notify"
)

type recordingNotifier struct {
	mu      sync.Mutex
	calls   []notify.DriftResult
	errOnce error
}

func (r *recordingNotifier) Notify(_ context.Context, res notify.DriftResult) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.errOnce != nil {
		err := r.errOnce
		r.errOnce = nil
		return err
	}
	r.calls = append(r.calls, res)
	return nil
}

func (r *recordingNotifier) Calls() []notify.DriftResult {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]notify.DriftResult, len(r.calls))
	copy(out, r.calls)
	return out
}

func bufferResult(service string, drifted bool) notify.DriftResult {
	return notify.DriftResult{ServiceName: service, HasDrift: drifted}
}

func TestBufferedNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewBufferedNotifier(nil, 2, 0)
}

func TestBufferedNotifier_FlushesAtCapacity(t *testing.T) {
	rec := &recordingNotifier{}
	b := notify.NewBufferedNotifier(rec, 3, 0)

	ctx := context.Background()
	_ = b.Notify(ctx, bufferResult("svc-a", false))
	_ = b.Notify(ctx, bufferResult("svc-b", true))

	if got := notify.BufferLen(b); got != 2 {
		t.Fatalf("expected 2 buffered, got %d", got)
	}
	if len(rec.Calls()) != 0 {
		t.Fatal("should not have flushed yet")
	}

	_ = b.Notify(ctx, bufferResult("svc-c", true))

	if got := notify.BufferLen(b); got != 0 {
		t.Fatalf("expected buffer empty after flush, got %d", got)
	}
	if calls := rec.Calls(); len(calls) != 1 {
		t.Fatalf("expected 1 flush call, got %d", len(calls))
	}
}

func TestBufferedNotifier_StopFlushesRemainder(t *testing.T) {
	rec := &recordingNotifier{}
	b := notify.NewBufferedNotifier(rec, 10, 0)

	ctx := context.Background()
	_ = b.Notify(ctx, bufferResult("svc-a", true))
	_ = b.Notify(ctx, bufferResult("svc-b", false))

	if err := b.Stop(ctx); err != nil {
		t.Fatalf("Stop returned error: %v", err)
	}
	if calls := rec.Calls(); len(calls) != 1 {
		t.Fatalf("expected 1 flush on Stop, got %d", len(calls))
	}
}

func TestBufferedNotifier_IntervalFlush(t *testing.T) {
	rec := &recordingNotifier{}
	b := notify.NewBufferedNotifier(rec, 100, 40*time.Millisecond)

	_ = b.Notify(context.Background(), bufferResult("svc-x", true))

	deadline := time.Now().Add(300 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(rec.Calls()) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	_ = b.Stop(context.Background())

	if len(rec.Calls()) == 0 {
		t.Fatal("expected interval-based flush to have fired")
	}
}

func TestBufferedNotifier_StopEmptyBufferNoError(t *testing.T) {
	rec := &recordingNotifier{}
	b := notify.NewBufferedNotifier(rec, 5, 0)
	if err := b.Stop(context.Background()); err != nil {
		t.Fatalf("unexpected error on empty Stop: %v", err)
	}
	if len(rec.Calls()) != 0 {
		t.Fatal("expected no calls on empty Stop")
	}
}

func TestBufferedNotifier_FlushErrorPropagates(t *testing.T) {
	sentinel := errors.New("notifier down")
	rec := &recordingNotifier{errOnce: sentinel}
	b := notify.NewBufferedNotifier(rec, 1, 0)

	err := b.Notify(context.Background(), bufferResult("svc-a", true))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
