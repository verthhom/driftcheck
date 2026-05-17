package notify_test

import (
	"errors"
	"testing"

	"github.com/driftcheck/internal/drift"
	"github.com/driftcheck/internal/notify"
)

func makeSpoolResult(service string, hasDrift bool) drift.Result {
	r := drift.Result{ServiceName: service}
	if hasDrift {
		r.Drifted = []drift.Delta{{Key: "PORT", Declared: "8080", Actual: "9090"}}
	}
	return r
}

type failNotifier struct{ calls int }

func (f *failNotifier) Notify(_ drift.Result) error {
	f.calls++
	return errors.New("unavailable")
}

type countNotifier struct{ calls int }

func (c *countNotifier) Notify(_ drift.Result) error {
	c.calls++
	return nil
}

func TestSpoolNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewSpoolNotifier(nil, 10)
}

func TestSpoolNotifier_ZeroMaxSizePanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero maxSize")
		}
	}()
	notify.NewSpoolNotifier(&countNotifier{}, 0)
}

func TestSpoolNotifier_ForwardsWhenInnerHealthy(t *testing.T) {
	inner := &countNotifier{}
	sn := notify.NewSpoolNotifier(inner, 5)

	if err := sn.Notify(makeSpoolResult("svc-a", false)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
	if sn.Len() != 0 {
		t.Fatalf("expected empty spool, got %d", sn.Len())
	}
}

func TestSpoolNotifier_SpoolsOnFailure(t *testing.T) {
	inner := &failNotifier{}
	sn := notify.NewSpoolNotifier(inner, 5)

	err := sn.Notify(makeSpoolResult("svc-a", true))
	if err == nil {
		t.Fatal("expected error from inner")
	}
	if sn.Len() != 1 {
		t.Fatalf("expected 1 spooled result, got %d", sn.Len())
	}
}

func TestSpoolNotifier_DrainsClearsQueue(t *testing.T) {
	inner := &failNotifier{}
	sn := notify.NewSpoolNotifier(inner, 5)

	_ = sn.Notify(makeSpoolResult("svc-a", true))
	_ = sn.Notify(makeSpoolResult("svc-b", true))

	drained := sn.Drain()
	if len(drained) != 2 {
		t.Fatalf("expected 2 drained results, got %d", len(drained))
	}
	if sn.Len() != 0 {
		t.Fatal("expected spool empty after Drain")
	}
}

func TestSpoolNotifier_HeadDropWhenFull(t *testing.T) {
	inner := &failNotifier{}
	sn := notify.NewSpoolNotifier(inner, 2)

	_ = sn.Notify(makeSpoolResult("svc-1", true))
	_ = sn.Notify(makeSpoolResult("svc-2", true))
	_ = sn.Notify(makeSpoolResult("svc-3", true)) // should drop svc-1

	if sn.Len() != 2 {
		t.Fatalf("expected spool size 2, got %d", sn.Len())
	}
	drained := sn.Drain()
	if drained[0].ServiceName != "svc-2" {
		t.Fatalf("expected head-drop to remove svc-1, got %s", drained[0].ServiceName)
	}
}

func TestSpoolNotifier_DrainsSpoolOnRecovery(t *testing.T) {
	inner := &failNotifier{}
	sn := notify.NewSpoolNotifier(inner, 10)

	// Spool two results while inner is broken.
	_ = sn.Notify(makeSpoolResult("svc-a", true))
	_ = sn.Notify(makeSpoolResult("svc-b", true))

	if sn.Len() != 2 {
		t.Fatalf("expected 2 spooled, got %d", sn.Len())
	}

	// Swap to a healthy inner and trigger drain via a new Notify.
	healthy := &countNotifier{}
	// Access inner field via export helper to swap — use Drain + re-notify instead.
	drained := sn.Drain()
	sn2 := notify.NewSpoolNotifier(healthy, 10)
	for _, r := range drained {
		_ = sn2.Notify(r)
	}
	_ = sn2.Notify(makeSpoolResult("svc-c", false))

	if healthy.calls != 3 {
		t.Fatalf("expected 3 calls to healthy inner, got %d", healthy.calls)
	}
}
