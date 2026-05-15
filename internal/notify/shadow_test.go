package notify_test

import (
	"context"
	"errors"
	"log"
	"strings"
	"sync/atomic"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

type countNotifier struct {
	calls int64
	err   error
}

func (c *countNotifier) Notify(_ context.Context, _ drift.Result) error {
	atomic.AddInt64(&c.calls, 1)
	return c.err
}

func makeShadowResult(service string, drifted bool) drift.Result {
	r := drift.Result{ServiceName: service}
	if drifted {
		r.Drifted = []drift.KeyDiff{{Key: "X", Declared: "a", Actual: "b"}}
	}
	return r
}

func TestShadowNotifier_NilPrimaryPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	notify.NewShadowNotifier(nil, &countNotifier{}, nil)
}

func TestShadowNotifier_NilShadowPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	notify.NewShadowNotifier(&countNotifier{}, nil, nil)
}

func TestShadowNotifier_CallsBoth(t *testing.T) {
	primary := &countNotifier{}
	shadow := &countNotifier{}
	sn := notify.NewShadowNotifier(primary, shadow, nil)

	if err := sn.Notify(context.Background(), makeShadowResult("svc", false)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	notify.ShadowWait(sn)

	if primary.calls != 1 {
		t.Errorf("primary calls = %d, want 1", primary.calls)
	}
	if shadow.calls != 1 {
		t.Errorf("shadow calls = %d, want 1", shadow.calls)
	}
}

func TestShadowNotifier_PrimaryErrorReturned(t *testing.T) {
	primary := &countNotifier{err: errors.New("primary fail")}
	shadow := &countNotifier{}
	sn := notify.NewShadowNotifier(primary, shadow, nil)

	err := sn.Notify(context.Background(), makeShadowResult("svc", true))
	notify.ShadowWait(sn)

	if err == nil || err.Error() != "primary fail" {
		t.Errorf("expected primary error, got %v", err)
	}
}

func TestShadowNotifier_ShadowErrorLogged(t *testing.T) {
	var buf strings.Builder
	logger := log.New(&buf, "[shadow] ", 0)

	primary := &countNotifier{}
	shadow := &countNotifier{err: errors.New("shadow fail")}
	sn := notify.NewShadowNotifier(primary, shadow, logger)

	if err := sn.Notify(context.Background(), makeShadowResult("alpha", true)); err != nil {
		t.Fatalf("unexpected primary error: %v", err)
	}
	notify.ShadowWait(sn)

	if !strings.Contains(buf.String(), "shadow fail") {
		t.Errorf("expected shadow error in log, got: %s", buf.String())
	}
	if !strings.Contains(buf.String(), "alpha") {
		t.Errorf("expected service name in log, got: %s", buf.String())
	}
}

func TestShadowNotifier_DefaultsToStderr(t *testing.T) {
	// Passing nil logger should not panic.
	primary := &countNotifier{}
	shadow := &countNotifier{err: errors.New("boom")}
	sn := notify.NewShadowNotifier(primary, shadow, nil)
	_ = sn.Notify(context.Background(), makeShadowResult("svc", false))
	notify.ShadowWait(sn)
}
