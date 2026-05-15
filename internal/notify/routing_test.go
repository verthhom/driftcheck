package notify_test

import (
	"context"
	"errors"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeRoutingResult(service string, drifted bool) drift.Result {
	r := drift.Result{ServiceName: service}
	if drifted {
		r.Drifted = []drift.Delta{{Key: "X", Expected: "a", Actual: "b"}}
	}
	return r
}

type captureNotifier struct {
	called int
	last   drift.Result
	err    error
}

func (c *captureNotifier) Notify(_ context.Context, r drift.Result) error {
	c.called++
	c.last = r
	return c.err
}

func TestRoutingNotifier_NilRulesPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic")
		}
	}()
	notify.NewRoutingNotifier(nil)
}

func TestRoutingNotifier_FirstMatchWins(t *testing.T) {
	a := &captureNotifier{}
	b := &captureNotifier{}

	n := notify.NewRoutingNotifier(nil,
		notify.RouteRule{Match: notify.OnlyDrifted, Notifier: a},
		notify.RouteRule{Match: func(drift.Result) bool { return true }, Notifier: b},
	)

	_ = n.Notify(context.Background(), makeRoutingResult("svc", true))

	if a.called != 1 {
		t.Fatalf("expected a called once, got %d", a.called)
	}
	if b.called != 0 {
		t.Fatalf("expected b not called, got %d", b.called)
	}
}

func TestRoutingNotifier_FallbackUsedWhenNoMatch(t *testing.T) {
	fb := &captureNotifier{}
	n := notify.NewRoutingNotifier(fb,
		notify.RouteRule{Match: notify.OnlyDrifted, Notifier: &captureNotifier{}},
	)

	_ = n.Notify(context.Background(), makeRoutingResult("svc", false))

	if fb.called != 1 {
		t.Fatalf("expected fallback called once, got %d", fb.called)
	}
}

func TestRoutingNotifier_NoMatchNoFallback_NoError(t *testing.T) {
	n := notify.NewRoutingNotifier(nil,
		notify.RouteRule{Match: notify.OnlyDrifted, Notifier: &captureNotifier{}},
	)

	if err := n.Notify(context.Background(), makeRoutingResult("svc", false)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRoutingNotifier_PropagatesInnerError(t *testing.T) {
	want := errors.New("boom")
	n := notify.NewRoutingNotifier(nil,
		notify.RouteRule{
			Match:    func(drift.Result) bool { return true },
			Notifier: &captureNotifier{err: want},
		},
	)

	got := n.Notify(context.Background(), makeRoutingResult("svc", true))
	if !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}

func TestRoutingNotifier_ServiceFilter(t *testing.T) {
	apiNotifier := &captureNotifier{}
	n := notify.NewRoutingNotifier(nil,
		notify.RouteRule{Match: notify.OnlyService("api"), Notifier: apiNotifier},
	)

	_ = n.Notify(context.Background(), makeRoutingResult("api", true))
	_ = n.Notify(context.Background(), makeRoutingResult("worker", true))

	if apiNotifier.called != 1 {
		t.Fatalf("expected api notifier called once, got %d", apiNotifier.called)
	}
}
