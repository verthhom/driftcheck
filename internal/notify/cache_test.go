package notify_test

import (
	"errors"
	"testing"
	"time"

	"driftcheck/internal/drift"
	"driftcheck/internal/notify"
)

func makeCacheResult(service string, drifted []drift.KeyDiff) drift.Result {
	return drift.Result{ServiceName: service, Drifted: drifted}
}

func TestCacheNotifier_NilInnerPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for nil inner")
		}
	}()
	notify.NewCacheNotifier(nil, time.Second)
}

func TestCacheNotifier_ZeroTTLPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic for zero ttl")
		}
	}()
	notify.NewCacheNotifier(&captureNotifier{}, 0)
}

func TestCacheNotifier_FirstCallForwarded(t *testing.T) {
	cap := &captureNotifier{}
	c := notify.NewCacheNotifier(cap, time.Minute)
	r := makeCacheResult("svc", nil)

	if err := c.Notify(r); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cap.count != 1 {
		t.Fatalf("expected 1 call, got %d", cap.count)
	}
}

func TestCacheNotifier_DuplicateSuppressed(t *testing.T) {
	cap := &captureNotifier{}
	c := notify.NewCacheNotifier(cap, time.Minute)
	r := makeCacheResult("svc", nil)

	_ = c.Notify(r)
	_ = c.Notify(r)

	if cap.count != 1 {
		t.Fatalf("expected 1 call after duplicate, got %d", cap.count)
	}
}

func TestCacheNotifier_ChangedResultForwarded(t *testing.T) {
	cap := &captureNotifier{}
	c := notify.NewCacheNotifier(cap, time.Minute)

	_ = c.Notify(makeCacheResult("svc", nil))
	_ = c.Notify(makeCacheResult("svc", []drift.KeyDiff{{Key: "FOO"}}))

	if cap.count != 2 {
		t.Fatalf("expected 2 calls for changed result, got %d", cap.count)
	}
}

func TestCacheNotifier_InvalidateForcesResend(t *testing.T) {
	cap := &captureNotifier{}
	c := notify.NewCacheNotifier(cap, time.Minute)
	r := makeCacheResult("svc", nil)

	_ = c.Notify(r)
	c.Invalidate("svc")
	_ = c.Notify(r)

	if cap.count != 2 {
		t.Fatalf("expected 2 calls after invalidate, got %d", cap.count)
	}
}

func TestCacheNotifier_TTLExpiry(t *testing.T) {
	cap := &captureNotifier{}
	c := notify.NewCacheNotifier(cap, time.Millisecond)
	r := makeCacheResult("svc", nil)

	_ = c.Notify(r)
	time.Sleep(5 * time.Millisecond)
	_ = c.Notify(r)

	if cap.count != 2 {
		t.Fatalf("expected 2 calls after TTL expiry, got %d", cap.count)
	}
}

func TestCacheNotifier_InnerErrorDoesNotCache(t *testing.T) {
	inner := &errorNotifier{err: errors.New("send failed")}
	c := notify.NewCacheNotifier(inner, time.Minute)
	r := makeCacheResult("svc", nil)

	_ = c.Notify(r)
	_ = c.Notify(r)

	if inner.count != 2 {
		t.Fatalf("expected 2 calls when inner errors, got %d", inner.count)
	}
}

func TestCacheNotifier_SeparateServicesIndependent(t *testing.T) {
	cap := &captureNotifier{}
	c := notify.NewCacheNotifier(cap, time.Minute)

	_ = c.Notify(makeCacheResult("svc-a", nil))
	_ = c.Notify(makeCacheResult("svc-b", nil))
	_ = c.Notify(makeCacheResult("svc-a", nil))

	if cap.count != 2 {
		t.Fatalf("expected 2 calls for two distinct services, got %d", cap.count)
	}
	if notify.CacheLen(c) != 2 {
		t.Fatalf("expected cache length 2, got %d", notify.CacheLen(c))
	}
}
