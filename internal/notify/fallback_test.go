package notify_test

import (
	"context"
	"errors"
	"log"
	"os"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

type stubNotifier struct {
	calls int
	err   error
}

func (s *stubNotifier) Notify(_ context.Context, _ drift.Result) error {
	s.calls++
	return s.err
}

func fallbackResult(service string, drifted bool) drift.Result {
	r := drift.Result{ServiceName: service}
	if drifted {
		r.Drifted = []drift.KeyDrift{{Key: "PORT", Declared: "8080", Live: "9090"}}
	}
	return r
}

func TestFallbackNotifier_NilPrimaryPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil primary")
		}
	}()
	notify.NewFallbackNotifier(nil, &stubNotifier{})
}

func TestFallbackNotifier_NoFallbacksPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic when no fallbacks provided")
		}
	}()
	notify.NewFallbackNotifier(&stubNotifier{})
}

func TestFallbackNotifier_PrimarySucceeds(t *testing.T) {
	primary := &stubNotifier{}
	fb := &stubNotifier{}
	n := notify.NewFallbackNotifier(primary, fb)

	if err := n.Notify(context.Background(), fallbackResult("svc", false)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if primary.calls != 1 {
		t.Errorf("expected primary called once, got %d", primary.calls)
	}
	if fb.calls != 0 {
		t.Errorf("expected fallback not called, got %d", fb.calls)
	}
}

func TestFallbackNotifier_FallbackCalledOnPrimaryError(t *testing.T) {
	primary := &stubNotifier{err: errors.New("primary down")}
	fb := &stubNotifier{}
	silent := log.New(os.Stderr, "", 0)
	n := notify.NewFallbackNotifier(primary, fb).WithLogger(silent)

	if err := n.Notify(context.Background(), fallbackResult("svc", true)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb.calls != 1 {
		t.Errorf("expected fallback called once, got %d", fb.calls)
	}
}

func TestFallbackNotifier_AllFail_ReturnsError(t *testing.T) {
	sentinel := errors.New("last error")
	primary := &stubNotifier{err: errors.New("primary down")}
	fb1 := &stubNotifier{err: errors.New("fb1 down")}
	fb2 := &stubNotifier{err: sentinel}
	silent := log.New(os.Stderr, "", 0)
	n := notify.NewFallbackNotifier(primary, fb1, fb2).WithLogger(silent)

	err := n.Notify(context.Background(), fallbackResult("svc", true))
	if err == nil {
		t.Fatal("expected error when all notifiers fail")
	}
	if !errors.Is(err, sentinel) {
		t.Errorf("expected last error wrapped, got: %v", err)
	}
}

func TestFallbackNotifier_SecondFallbackSucceeds(t *testing.T) {
	primary := &stubNotifier{err: errors.New("primary down")}
	fb1 := &stubNotifier{err: errors.New("fb1 down")}
	fb2 := &stubNotifier{}
	silent := log.New(os.Stderr, "", 0)
	n := notify.NewFallbackNotifier(primary, fb1, fb2).WithLogger(silent)

	if err := n.Notify(context.Background(), fallbackResult("svc", true)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fb2.calls != 1 {
		t.Errorf("expected second fallback called once, got %d", fb2.calls)
	}
}
