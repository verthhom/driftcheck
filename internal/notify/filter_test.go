package notify_test

import (
	"context"
	"errors"
	"testing"

	"github.com/example/driftcheck/internal/drift"
	"github.com/example/driftcheck/internal/notify"
)

type recordingNotifier struct {
	calls []drift.Result
	err   error
}

func (r *recordingNotifier) Notify(_ context.Context, result drift.Result) error {
	r.calls = append(r.calls, result)
	return r.err
}

func noDriftResult(service string) drift.Result {
	return drift.Result{ServiceName: service, Drifted: nil}
}

func withDriftResult(service string) drift.Result {
	return drift.Result{
		ServiceName: service,
		Drifted:     []drift.Delta{{Key: "FOO", Declared: "a", Actual: "b"}},
	}
}

func TestFilteredNotifier_NilInner(t *testing.T) {
	_, err := notify.NewFilteredNotifier(nil)
	if !errors.Is(err, notify.ErrNilInner) {
		t.Fatalf("expected ErrNilInner, got %v", err)
	}
}

func TestFilteredNotifier_NoFilters_AlwaysForwards(t *testing.T) {
	rec := &recordingNotifier{}
	fn, _ := notify.NewFilteredNotifier(rec)

	_ = fn.Notify(context.Background(), noDriftResult("svc"))
	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(rec.calls))
	}
}

func TestFilteredNotifier_OnlyDrifted_Suppresses(t *testing.T) {
	rec := &recordingNotifier{}
	fn, _ := notify.NewFilteredNotifier(rec, notify.OnlyDrifted)

	_ = fn.Notify(context.Background(), noDriftResult("svc"))
	if len(rec.calls) != 0 {
		t.Fatalf("expected 0 calls for no-drift result, got %d", len(rec.calls))
	}
}

func TestFilteredNotifier_OnlyDrifted_Passes(t *testing.T) {
	rec := &recordingNotifier{}
	fn, _ := notify.NewFilteredNotifier(rec, notify.OnlyDrifted)

	_ = fn.Notify(context.Background(), withDriftResult("svc"))
	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 call for drifted result, got %d", len(rec.calls))
	}
}

func TestFilteredNotifier_OnlyService_Filters(t *testing.T) {
	rec := &recordingNotifier{}
	fn, _ := notify.NewFilteredNotifier(rec, notify.OnlyService("payments"))

	_ = fn.Notify(context.Background(), withDriftResult("auth"))
	_ = fn.Notify(context.Background(), withDriftResult("payments"))

	if len(rec.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(rec.calls))
	}
	if rec.calls[0].ServiceName != "payments" {
		t.Fatalf("expected payments, got %s", rec.calls[0].ServiceName)
	}
}

func TestFilteredNotifier_InnerError_Propagates(t *testing.T) {
	sentinel := errors.New("send failed")
	rec := &recordingNotifier{err: sentinel}
	fn, _ := notify.NewFilteredNotifier(rec)

	err := fn.Notify(context.Background(), withDriftResult("svc"))
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}
