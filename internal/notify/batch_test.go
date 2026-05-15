package notify_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

type batchResult = drift.Result

type recordingNotifier struct {
	mu      sync.Mutex
	calls   []drift.Result
	errOnce error
}

func (r *recordingNotifier) Notify(_ context.Context, res drift.Result) error {
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

func makeBatchResult(service string, keys ...string) batchResult {
	diffs := make([]drift.Diff, len(keys))
	for i, k := range keys {
		diffs[i] = drift.Diff{Key: k, Declared: "a", Actual: "b"}
	}
	return batchResult{ServiceName: service, Diffs: diffs}
}

func TestBatchNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewBatchNotifier(nil, time.Second)
}

func TestBatchNotifier_ZeroWindowForwardsImmediately(t *testing.T) {
	rec := &recordingNotifier{}
	bn := notify.NewBatchNotifier(rec, 0)
	res := makeBatchResult("svc", "KEY")
	if err := bn.Notify(context.Background(), res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(rec.calls))
	}
}

func TestBatchNotifier_BatchesThenFlushes(t *testing.T) {
	rec := &recordingNotifier{}
	bn := notify.NewBatchNotifier(rec, 10*time.Second) // long window — manual flush

	_ = bn.Notify(context.Background(), makeBatchResult("alpha", "A"))
	_ = bn.Notify(context.Background(), makeBatchResult("beta", "B"))

	rec.mu.Lock()
	if len(rec.calls) != 0 {
		rec.mu.Unlock()
		t.Fatal("expected no calls before flush")
	}
	rec.mu.Unlock()

	if err := bn.Flush(context.Background()); err != nil {
		t.Fatalf("flush error: %v", err)
	}

	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 combined call after flush, got %d", len(rec.calls))
	}
	if len(rec.calls[0].Diffs) != 2 {
		t.Fatalf("expected 2 combined diffs, got %d", len(rec.calls[0].Diffs))
	}
}

func TestBatchNotifier_FlushEmptyIsNoop(t *testing.T) {
	rec := &recordingNotifier{}
	bn := notify.NewBatchNotifier(rec, 5*time.Second)
	if err := bn.Flush(context.Background()); err != nil {
		t.Fatalf("unexpected error on empty flush: %v", err)
	}
	rec.mu.Lock()
	defer rec.mu.Unlock()
	if len(rec.calls) != 0 {
		t.Fatalf("expected no calls, got %d", len(rec.calls))
	}
}

func TestBatchNotifier_FlushPropagatesError(t *testing.T) {
	sentinel := errors.New("downstream failure")
	rec := &recordingNotifier{errOnce: sentinel}
	bn := notify.NewBatchNotifier(rec, 5*time.Second)
	_ = bn.Notify(context.Background(), makeBatchResult("svc", "X"))
	if err := bn.Flush(context.Background()); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestBatchNotifier_AutoFlushViaTimer(t *testing.T) {
	rec := &recordingNotifier{}
	bn := notify.NewBatchNotifier(rec, 50*time.Millisecond)
	_ = bn.Notify(context.Background(), makeBatchResult("svc", "K"))

	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		rec.mu.Lock()
		n := len(rec.calls)
		rec.mu.Unlock()
		if n > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("timer did not auto-flush within deadline")
}
