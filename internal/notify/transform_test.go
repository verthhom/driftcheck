package notify_test

import (
	"errors"
	"testing"

	"github.com/driftcheck/internal/drift"
	"github.com/driftcheck/internal/notify"
)

// captureNotifier records the last Result it received.
type captureNotifier struct {
	last drift.Result
	err  error
}

func (c *captureNotifier) Notify(r drift.Result) error {
	c.last = r
	return c.err
}

func makeTransformResult(service string, diffs []drift.Difference) drift.Result {
	return drift.Result{Service: service, Diffs: diffs}
}

func TestTransformingNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewTransformingNotifier(nil)
}

func TestTransformingNotifier_NoTransformers_Passthrough(t *testing.T) {
	cap := &captureNotifier{}
	tn := notify.NewTransformingNotifier(cap)
	r := makeTransformResult("svc", nil)
	if err := tn.Notify(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.last.Service != "svc" {
		t.Errorf("service = %q, want %q", cap.last.Service, "svc")
	}
}

func TestTransformingNotifier_ForwardsInnerError(t *testing.T) {
	want := errors.New("boom")
	cap := &captureNotifier{err: want}
	tn := notify.NewTransformingNotifier(cap)
	if got := tn.Notify(makeTransformResult("svc", nil)); !errors.Is(got, want) {
		t.Errorf("error = %v, want %v", got, want)
	}
}

func TestRedactKeys_RedactsMatchingKeys(t *testing.T) {
	cap := &captureNotifier{}
	tn := notify.NewTransformingNotifier(cap, notify.RedactKeys("secret", "password"))
	diffs := []drift.Difference{
		{Key: "DB_PASSWORD", Want: "hunter2", Got: "letmein"},
		{Key: "APP_PORT", Want: "8080", Got: "9090"},
		{Key: "API_SECRET_KEY", Want: "abc", Got: "xyz"},
	}
	if err := tn.Notify(makeTransformResult("svc", diffs)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range cap.last.Diffs {
		switch d.Key {
		case "DB_PASSWORD", "API_SECRET_KEY":
			if d.Got != "[REDACTED]" || d.Want != "[REDACTED]" {
				t.Errorf("key %q not redacted: got=%q want=%q", d.Key, d.Got, d.Want)
			}
		case "APP_PORT":
			if d.Got == "[REDACTED]" {
				t.Errorf("key %q should not be redacted", d.Key)
			}
		}
	}
}

func TestPrefixService_PrependsPrefixToServiceName(t *testing.T) {
	cap := &captureNotifier{}
	tn := notify.NewTransformingNotifier(cap, notify.PrefixService("prod-"))
	if err := tn.Notify(makeTransformResult("payments", nil)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := cap.last.Service, "prod-payments"; got != want {
		t.Errorf("service = %q, want %q", got, want)
	}
}

func TestTransformingNotifier_AppliesTransformersInOrder(t *testing.T) {
	cap := &captureNotifier{}
	tn := notify.NewTransformingNotifier(cap,
		notify.PrefixService("a-"),
		notify.PrefixService("b-"),
	)
	if err := tn.Notify(makeTransformResult("svc", nil)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got, want := cap.last.Service, "b-a-svc"; got != want {
		t.Errorf("service = %q, want %q", got, want)
	}
}
