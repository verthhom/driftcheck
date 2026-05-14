package notify_test

import (
	"errors"
	"sync/atomic"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeDedupeDrift(service string, keys ...string) drift.Result {
	diffs := make([]drift.Diff, 0, len(keys))
	for _, k := range keys {
		diffs = append(diffs, drift.Diff{Key: k, Declared: "a", Actual: "b"})
	}
	return drift.Result{ServiceName: service, Diffs: diffs}
}

type countingNotifier struct {
	calls atomic.Int32
	err   error
}

func (c *countingNotifier) Notify(_ drift.Result) error {
	c.calls.Add(1)
	return c.err
}

func TestDedupeNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewDedupeNotifier(nil)
}

func TestDedupeNotifier_ForwardsNoDrift(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner)

	r := drift.Result{ServiceName: "svc", Diffs: nil}
	if err := d.Notify(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := d.Notify(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls for no-drift, got %d", got)
	}
}

func TestDedupeNotifier_SuppressesDuplicate(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner)

	r := makeDedupeDrift("svc", "KEY_A")
	_ = d.Notify(r)
	_ = d.Notify(r)
	_ = d.Notify(r)

	if got := inner.calls.Load(); got != 1 {
		t.Fatalf("expected 1 call, got %d", got)
	}
}

func TestDedupeNotifier_DifferentServicesNotSuppressed(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner)

	_ = d.Notify(makeDedupeDrift("svc-a", "KEY_A"))
	_ = d.Notify(makeDedupeDrift("svc-b", "KEY_A"))

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls, got %d", got)
	}
}

func TestDedupeNotifier_ResetAllowsResend(t *testing.T) {
	inner := &countingNotifier{}
	d := notify.NewDedupeNotifier(inner)

	r := makeDedupeDrift("svc", "KEY_A")
	_ = d.Notify(r)
	d.Reset()
	_ = d.Notify(r)

	if got := inner.calls.Load(); got != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", got)
	}
}

func TestDedupeNotifier_PropagatesError(t *testing.T) {
	want := errors.New("send failed")
	inner := &countingNotifier{err: want}
	d := notify.NewDedupeNotifier(inner)

	if err := d.Notify(makeDedupeDrift("svc", "KEY_X")); !errors.Is(err, want) {
		t.Fatalf("expected %v, got %v", want, err)
	}
}
