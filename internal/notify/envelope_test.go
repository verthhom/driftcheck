package notify_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"driftcheck/internal/notify"
)

type captureNotifier struct {
	received notify.Result
	err      error
}

func (c *captureNotifier) Notify(r notify.Result) error {
	c.received = r
	return c.err
}

func makeEnvelopeResult(service string) notify.Result {
	return notify.Result{
		Service: service,
		Drifted: []string{"KEY_A"},
	}
}

func TestEnvelopeNotifier_NilInnerPanics(t *testing.T) {
	t.Parallel()
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewEnvelopeNotifier(nil, notify.Envelope{})
}

func TestEnvelopeNotifier_StampsEnvironment(t *testing.T) {
	t.Parallel()
	cap := &captureNotifier{}
	n := notify.NewEnvelopeNotifier(cap, notify.Envelope{Environment: "production"})
	if err := n.Notify(makeEnvelopeResult("svc")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.received.Meta["environment"] != "production" {
		t.Errorf("expected environment=production, got %q", cap.received.Meta["environment"])
	}
}

func TestEnvelopeNotifier_StampsHost(t *testing.T) {
	t.Parallel()
	cap := &captureNotifier{}
	n := notify.NewEnvelopeNotifier(cap, notify.Envelope{Host: "node-42"})
	_ = n.Notify(makeEnvelopeResult("svc"))
	if cap.received.Meta["host"] != "node-42" {
		t.Errorf("expected host=node-42, got %q", cap.received.Meta["host"])
	}
}

func TestEnvelopeNotifier_StampsTimestamp(t *testing.T) {
	t.Parallel()
	cap := &captureNotifier{}
	fixed := time.Unix(1_700_000_000, 0)
	n := notify.NewEnvelopeNotifier(cap, notify.Envelope{})
	notify.EnvelopeNotifierSetNow(n, func() time.Time { return fixed })
	_ = n.Notify(makeEnvelopeResult("svc"))
	want := fmt.Sprintf("%d", fixed.Unix())
	if cap.received.Meta["envelope_ts"] != want {
		t.Errorf("expected envelope_ts=%s, got %q", want, cap.received.Meta["envelope_ts"])
	}
}

func TestEnvelopeNotifier_PreservesExistingMeta(t *testing.T) {
	t.Parallel()
	cap := &captureNotifier{}
	n := notify.NewEnvelopeNotifier(cap, notify.Envelope{Environment: "staging"})
	r := makeEnvelopeResult("svc")
	r.Meta = map[string]string{"custom": "value"}
	_ = n.Notify(r)
	if cap.received.Meta["custom"] != "value" {
		t.Errorf("existing meta key was overwritten")
	}
	if cap.received.Meta["environment"] != "staging" {
		t.Errorf("envelope environment not set")
	}
}

func TestEnvelopeNotifier_ForwardsInnerError(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("inner failure")
	cap := &captureNotifier{err: sentinel}
	n := notify.NewEnvelopeNotifier(cap, notify.Envelope{Environment: "test"})
	if err := n.Notify(makeEnvelopeResult("svc")); !errors.Is(err, sentinel) {
		t.Errorf("expected sentinel error, got %v", err)
	}
}

func TestEnvelopeNotifier_EmptyEnvelopeOnlyStampsTimestamp(t *testing.T) {
	t.Parallel()
	cap := &captureNotifier{}
	n := notify.NewEnvelopeNotifier(cap, notify.Envelope{})
	_ = n.Notify(makeEnvelopeResult("svc"))
	if _, ok := cap.received.Meta["environment"]; ok {
		t.Error("environment key should not be set for empty envelope")
	}
	if _, ok := cap.received.Meta["envelope_ts"]; !ok {
		t.Error("envelope_ts should always be set")
	}
}
