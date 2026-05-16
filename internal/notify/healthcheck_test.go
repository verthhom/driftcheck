package notify_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeHealthResult(service string, drifted bool) drift.Result {
	r := drift.Result{ServiceName: service}
	if drifted {
		r.Drifted = []drift.KeyDiff{{Key: "PORT", Expected: "8080", Actual: "9090"}}
	}
	return r
}

func TestHealthCheckNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewHealthCheckNotifier(nil)
}

func TestHealthCheckNotifier_InitiallyHealthy(t *testing.T) {
	inner := &fakeNotifier{}
	h := notify.NewHealthCheckNotifier(inner)
	if !h.Health().Healthy {
		t.Fatal("expected initial state to be healthy")
	}
}

func TestHealthCheckNotifier_RecordsSuccess(t *testing.T) {
	inner := &fakeNotifier{}
	h := notify.NewHealthCheckNotifier(inner)

	before := time.Now()
	_ = h.Notify(context.Background(), makeHealthResult("svc", false))
	after := time.Now()

	s := h.Health()
	if !s.Healthy {
		t.Fatal("expected healthy after successful notify")
	}
	if s.LastSuccess.Before(before) || s.LastSuccess.After(after) {
		t.Errorf("LastSuccess %v not in expected range", s.LastSuccess)
	}
	if s.FailureCount != 0 {
		t.Errorf("expected FailureCount 0, got %d", s.FailureCount)
	}
}

func TestHealthCheckNotifier_RecordsFailure(t *testing.T) {
	inner := &fakeNotifier{err: errors.New("send failed")}
	h := notify.NewHealthCheckNotifier(inner)

	_ = h.Notify(context.Background(), makeHealthResult("svc", true))

	s := h.Health()
	if s.Healthy {
		t.Fatal("expected unhealthy after failed notify")
	}
	if s.FailureCount != 1 {
		t.Errorf("expected FailureCount 1, got %d", s.FailureCount)
	}
	if s.Message == "" {
		t.Error("expected non-empty Message after failure")
	}
}

func TestHealthCheckNotifier_IncrementsFailureCount(t *testing.T) {
	inner := &fakeNotifier{err: errors.New("oops")}
	h := notify.NewHealthCheckNotifier(inner)

	for i := 0; i < 3; i++ {
		_ = h.Notify(context.Background(), makeHealthResult("svc", true))
	}

	if h.Health().FailureCount != 3 {
		t.Errorf("expected FailureCount 3, got %d", h.Health().FailureCount)
	}
}

func TestHealthCheckNotifier_ResetClearsState(t *testing.T) {
	inner := &fakeNotifier{err: errors.New("oops")}
	h := notify.NewHealthCheckNotifier(inner)
	_ = h.Notify(context.Background(), makeHealthResult("svc", true))

	h.Reset()

	s := h.Health()
	if !s.Healthy {
		t.Fatal("expected healthy after Reset")
	}
	if s.FailureCount != 0 {
		t.Errorf("expected FailureCount 0 after Reset, got %d", s.FailureCount)
	}
}

func TestHealthCheckNotifier_ForwardsError(t *testing.T) {
	want := errors.New("downstream error")
	inner := &fakeNotifier{err: want}
	h := notify.NewHealthCheckNotifier(inner)

	got := h.Notify(context.Background(), makeHealthResult("svc", true))
	if !errors.Is(got, want) {
		t.Errorf("expected %v, got %v", want, got)
	}
}
