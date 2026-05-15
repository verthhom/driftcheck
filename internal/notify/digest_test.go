package notify_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

type digestCapture struct {
	results []drift.Result
	err     error
}

func (c *digestCapture) Notify(r drift.Result) error {
	c.results = append(c.results, r)
	return c.err
}

func makeDigestResult(service string, keys ...string) drift.Result {
	var drifted []drift.KeyDiff
	for _, k := range keys {
		drifted = append(drifted, drift.KeyDiff{Key: k, Declared: "a", Deployed: "b"})
	}
	return drift.Result{ServiceName: service, Drifted: drifted}
}

func TestDigestNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewDigestNotifier(nil, 0)
}

func TestDigestNotifier_NotifyQueuesResult(t *testing.T) {
	cap := &digestCapture{}
	d := notify.NewDigestNotifier(cap, 0)

	_ = d.Notify(makeDigestResult("svc-a", "PORT"))
	_ = d.Notify(makeDigestResult("svc-b", "HOST"))

	if got := notify.DigestLen(d); got != 2 {
		t.Fatalf("expected 2 queued results, got %d", got)
	}
	if len(cap.results) != 0 {
		t.Fatal("inner should not be called before Flush")
	}
}

func TestDigestNotifier_FlushCombinesResults(t *testing.T) {
	cap := &digestCapture{}
	d := notify.NewDigestNotifier(cap, 0)

	_ = d.Notify(makeDigestResult("svc-a", "PORT"))
	_ = d.Notify(makeDigestResult("svc-b", "HOST", "DB"))

	if err := d.Flush(); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}

	if len(cap.results) != 1 {
		t.Fatalf("expected 1 digest result, got %d", len(cap.results))
	}
	r := cap.results[0]
	if !strings.Contains(r.ServiceName, "svc-a") || !strings.Contains(r.ServiceName, "svc-b") {
		t.Errorf("digest service name missing expected services: %q", r.ServiceName)
	}
	if len(r.Drifted) != 3 {
		t.Errorf("expected 3 drifted keys in digest, got %d", len(r.Drifted))
	}
	if notify.DigestLen(d) != 0 {
		t.Error("queue should be empty after flush")
	}
}

func TestDigestNotifier_FlushEmptyIsNoop(t *testing.T) {
	cap := &digestCapture{}
	d := notify.NewDigestNotifier(cap, 0)

	if err := d.Flush(); err != nil {
		t.Fatalf("flush on empty queue should not error: %v", err)
	}
	if len(cap.results) != 0 {
		t.Fatal("inner should not be called for empty flush")
	}
}

func TestDigestNotifier_FlushPropagatesInnerError(t *testing.T) {
	sentinel := errors.New("inner failed")
	cap := &digestCapture{err: sentinel}
	d := notify.NewDigestNotifier(cap, 0)

	_ = d.Notify(makeDigestResult("svc-a"))
	if err := d.Flush(); !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
}

func TestDigestNotifier_AutoFlushesAfterWindow(t *testing.T) {
	cap := &digestCapture{}
	d := notify.NewDigestNotifier(cap, 50*time.Millisecond)

	_ = d.Notify(makeDigestResult("svc-a", "KEY"))

	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if len(cap.results) > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if len(cap.results) == 0 {
		t.Fatal("expected auto-flush to fire within window")
	}
}
