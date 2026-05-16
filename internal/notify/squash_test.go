package notify_test

import (
	"errors"
	"testing"

	"github.com/driftcheck/internal/drift"
	"github.com/driftcheck/internal/notify"
)

func makeSquashResult(service string, drifted bool) drift.Result {
	r := drift.Result{ServiceName: service}
	if drifted {
		r.Drifted = []drift.KeyDiff{{Key: "ENV", Expected: "a", Actual: "b"}}
	}
	return r
}

func TestSquashNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewSquashNotifier(nil)
}

func TestSquashNotifier_EmptyServiceNameErrors(t *testing.T) {
	cap := &capturingNotifier{}
	sq := notify.NewSquashNotifier(cap)
	err := sq.Notify(drift.Result{})
	if err == nil {
		t.Fatal("expected error for empty service name")
	}
}

func TestSquashNotifier_SquashesMultipleCalls(t *testing.T) {
	cap := &capturingNotifier{}
	sq := notify.NewSquashNotifier(cap)

	_ = sq.Notify(makeSquashResult("svc", false))
	_ = sq.Notify(makeSquashResult("svc", false))
	_ = sq.Notify(makeSquashResult("svc", true)) // latest

	if sq.Len() != 1 {
		t.Fatalf("expected 1 pending, got %d", sq.Len())
	}
	if err := sq.Flush(); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}
	if len(cap.results) != 1 {
		t.Fatalf("expected 1 forwarded result, got %d", len(cap.results))
	}
	if !cap.results[0].HasDrift() {
		t.Error("expected latest (drifted) result to be forwarded")
	}
}

func TestSquashNotifier_FlushClearsQueue(t *testing.T) {
	cap := &capturingNotifier{}
	sq := notify.NewSquashNotifier(cap)

	_ = sq.Notify(makeSquashResult("svc", false))
	_ = sq.Flush()

	if sq.Len() != 0 {
		t.Fatalf("expected empty queue after flush, got %d", sq.Len())
	}
}

func TestSquashNotifier_MultipleServicesKeptSeparate(t *testing.T) {
	cap := &capturingNotifier{}
	sq := notify.NewSquashNotifier(cap)

	_ = sq.Notify(makeSquashResult("alpha", false))
	_ = sq.Notify(makeSquashResult("beta", true))

	if sq.Len() != 2 {
		t.Fatalf("expected 2 pending, got %d", sq.Len())
	}
	_ = sq.Flush()
	if len(cap.results) != 2 {
		t.Fatalf("expected 2 forwarded results, got %d", len(cap.results))
	}
}

func TestSquashNotifier_FlushPropagatesInnerError(t *testing.T) {
	failing := &failingNotifier{err: errors.New("send failed")}
	sq := notify.NewSquashNotifier(failing)

	_ = sq.Notify(makeSquashResult("svc", true))
	err := sq.Flush()
	if err == nil {
		t.Fatal("expected error from failing inner notifier")
	}
}
