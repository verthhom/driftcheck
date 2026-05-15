package notify_test

import (
	"context"
	"errors"
	"testing"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

type captureNotifier struct {
	called bool
	result drift.Result
	err    error
}

func (c *captureNotifier) Notify(_ context.Context, r drift.Result) error {
	c.called = true
	c.result = r
	return c.err
}

func makePriorityResult(service string, driftedKeys []string) drift.Result {
	drifted := make([]drift.Difference, len(driftedKeys))
	for i, k := range driftedKeys {
		drifted[i] = drift.Difference{Key: k, Expected: "a", Actual: "b"}
	}
	return drift.Result{ServiceName: service, Drifted: drifted}
}

func TestPriorityNotifier_NilHighPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil high notifier")
		}
	}()
	notify.NewPriorityNotifier(2, nil, &captureNotifier{})
}

func TestPriorityNotifier_NilLowPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil low notifier")
		}
	}()
	notify.NewPriorityNotifier(2, &captureNotifier{}, nil)
}

func TestPriorityNotifier_ZeroThresholdPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero threshold")
		}
	}()
	notify.NewPriorityNotifier(0, &captureNotifier{}, &captureNotifier{})
}

func TestPriorityNotifier_BelowThresholdUsesLow(t *testing.T) {
	high := &captureNotifier{}
	low := &captureNotifier{}
	p := notify.NewPriorityNotifier(3, high, low)

	result := makePriorityResult("svc", []string{"KEY_A"})
	if err := p.Notify(context.Background(), result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if high.called {
		t.Error("high notifier should not have been called")
	}
	if !low.called {
		t.Error("low notifier should have been called")
	}
}

func TestPriorityNotifier_AtThresholdUsesHigh(t *testing.T) {
	high := &captureNotifier{}
	low := &captureNotifier{}
	p := notify.NewPriorityNotifier(2, high, low)

	result := makePriorityResult("svc", []string{"KEY_A", "KEY_B"})
	if err := p.Notify(context.Background(), result); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !high.called {
		t.Error("high notifier should have been called")
	}
	if low.called {
		t.Error("low notifier should not have been called")
	}
}

func TestPriorityNotifier_HighErrorWrapped(t *testing.T) {
	high := &captureNotifier{err: errors.New("upstream down")}
	low := &captureNotifier{}
	p := notify.NewPriorityNotifier(1, high, low)

	result := makePriorityResult("svc", []string{"KEY_A"})
	err := p.Notify(context.Background(), result)
	if err == nil {
		t.Fatal("expected error from high notifier")
	}
	if !errors.Is(err, high.err) {
		t.Errorf("expected wrapped error, got: %v", err)
	}
}
