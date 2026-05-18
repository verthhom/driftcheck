package notify_test

import (
	"context"
	"errors"
	"testing"

	"github.com/driftcheck/driftcheck/internal/drift"
	"github.com/driftcheck/driftcheck/internal/notify"
)

func makeScopeResult(service string, drifted bool) drift.Result {
	r := drift.Result{Service: service}
	if drifted {
		r.Drifted = []drift.Delta{{Key: "PORT", Expected: "8080", Actual: "9090"}}
	}
	return r
}

type scopeCapture struct {
	called []drift.Result
	err    error
}

func (c *scopeCapture) Notify(_ context.Context, r drift.Result) error {
	c.called = append(c.called, r)
	return c.err
}

func TestScopeNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewScopeNotifier(nil, "svc-a")
}

func TestScopeNotifier_NoScopesPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for empty scopes")
		}
	}()
	notify.NewScopeNotifier(&scopeCapture{})
}

func TestScopeNotifier_InScopeForwards(t *testing.T) {
	cap := &scopeCapture{}
	n := notify.NewScopeNotifier(cap, "svc-a", "svc-b")

	r := makeScopeResult("svc-a", true)
	if err := n.Notify(context.Background(), r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.called) != 1 {
		t.Fatalf("expected 1 forwarded call, got %d", len(cap.called))
	}
}

func TestScopeNotifier_OutOfScopeDropped(t *testing.T) {
	cap := &scopeCapture{}
	n := notify.NewScopeNotifier(cap, "svc-a")

	if err := n.Notify(context.Background(), makeScopeResult("svc-z", true)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cap.called) != 0 {
		t.Fatalf("expected 0 forwarded calls, got %d", len(cap.called))
	}
}

func TestScopeNotifier_ForwardsInnerError(t *testing.T) {
	sentinel := errors.New("inner failure")
	cap := &scopeCapture{err: sentinel}
	n := notify.NewScopeNotifier(cap, "svc-a")

	err := n.Notify(context.Background(), makeScopeResult("svc-a", false))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestScopeNotifier_ScopesReturnsAll(t *testing.T) {
	cap := &scopeCapture{}
	n := notify.NewScopeNotifier(cap, "alpha", "beta", "gamma")

	scopes := n.Scopes()
	if len(scopes) != 3 {
		t.Fatalf("expected 3 scopes, got %d", len(scopes))
	}

	scopeMap := notify.ScopeMap(n)
	for _, s := range []string{"alpha", "beta", "gamma"} {
		if _, ok := scopeMap[s]; !ok {
			t.Errorf("scope %q missing from map", s)
		}
	}
}

func TestScopeNotifier_MultipleInScope(t *testing.T) {
	cap := &scopeCapture{}
	n := notify.NewScopeNotifier(cap, "svc-a", "svc-b")

	for _, svc := range []string{"svc-a", "svc-b", "svc-c", "svc-a"} {
		_ = n.Notify(context.Background(), makeScopeResult(svc, false))
	}
	// svc-a (x2) + svc-b = 3 forwarded; svc-c dropped
	if len(cap.called) != 3 {
		t.Fatalf("expected 3 forwarded calls, got %d", len(cap.called))
	}
}
