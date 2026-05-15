package notify_test

import (
	"context"
	"errors"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeReplayResult(service string, hasDrift bool) drift.Result {
	r := drift.Result{ServiceName: service}
	if hasDrift {
		r.Drifted = []drift.KeyDiff{{Key: "ENV", Expected: "a", Actual: "b"}}
	}
	return r
}

func TestReplayNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewReplayNotifier(nil, 5)
}

func TestReplayNotifier_ZeroCapacityPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero capacity")
		}
	}()
	notify.NewReplayNotifier(&stubNotifier{}, 0)
}

func TestReplayNotifier_ForwardsToInner(t *testing.T) {
	stub := &stubNotifier{}
	rn := notify.NewReplayNotifier(stub, 10)
	res := makeReplayResult("svc", false)

	if err := rn.Notify(context.Background(), res); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", stub.calls)
	}
}

func TestReplayNotifier_BuffersResults(t *testing.T) {
	rn := notify.NewReplayNotifier(&stubNotifier{}, 3)
	for i := 0; i < 3; i++ {
		_ = rn.Notify(context.Background(), makeReplayResult("svc", false))
	}
	if notify.ReplayBuf(rn) != 3 {
		t.Fatalf("expected buffer len 3, got %d", notify.ReplayBuf(rn))
	}
}

func TestReplayNotifier_EvictsOldest(t *testing.T) {
	rn := notify.NewReplayNotifier(&stubNotifier{}, 2)
	for i := 0; i < 4; i++ {
		_ = rn.Notify(context.Background(), makeReplayResult("svc", false))
	}
	if notify.ReplayBuf(rn) != 2 {
		t.Fatalf("expected buffer len 2 after eviction, got %d", notify.ReplayBuf(rn))
	}
}

func TestReplayNotifier_ReplayCallsTarget(t *testing.T) {
	rn := notify.NewReplayNotifier(&stubNotifier{}, 5)
	for i := 0; i < 3; i++ {
		_ = rn.Notify(context.Background(), makeReplayResult("svc", true))
	}

	target := &stubNotifier{}
	if err := rn.Replay(context.Background(), target); err != nil {
		t.Fatalf("unexpected replay error: %v", err)
	}
	if target.calls != 3 {
		t.Fatalf("expected 3 replay calls, got %d", target.calls)
	}
}

func TestReplayNotifier_ReplayCollectsErrors(t *testing.T) {
	rn := notify.NewReplayNotifier(&stubNotifier{}, 5)
	_ = rn.Notify(context.Background(), makeReplayResult("svc", false))

	failing := &stubNotifier{err: errors.New("send failed")}
	err := rn.Replay(context.Background(), failing)
	if err == nil {
		t.Fatal("expected error from failing target, got nil")
	}
}
