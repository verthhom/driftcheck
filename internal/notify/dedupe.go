package notify

import (
	"crypto/sha256"
	"fmt"
	"sync"

	"driftcheck/internal/drift"
)

// DedupeNotifier wraps a Notifier and suppresses duplicate alerts within a
// session. Two results are considered duplicates when their service name and
// the set of drifted keys produce the same fingerprint.
type DedupeNotifier struct {
	inner  Notifier
	mu     sync.Mutex
	seen   map[string]struct{}
}

// NewDedupeNotifier returns a DedupeNotifier that forwards each unique
// drift fingerprint to inner exactly once per process lifetime.
func NewDedupeNotifier(inner Notifier) *DedupeNotifier {
	if inner == nil {
		panic("dedupe: inner notifier must not be nil")
	}
	return &DedupeNotifier{
		inner: inner,
		seen:  make(map[string]struct{}),
	}
}

// Notify forwards r to the inner notifier only when the fingerprint of r has
// not been seen before. Calls with no drift are always forwarded.
func (d *DedupeNotifier) Notify(r drift.Result) error {
	if !r.HasDrift() {
		return d.inner.Notify(r)
	}

	fp := fingerprint(r)

	d.mu.Lock()
	_, already := d.seen[fp]
	if !already {
		d.seen[fp] = struct{}{}
	}
	d.mu.Unlock()

	if already {
		return nil
	}
	return d.inner.Notify(r)
}

// Reset clears all remembered fingerprints, allowing previously suppressed
// results to be forwarded again.
func (d *DedupeNotifier) Reset() {
	d.mu.Lock()
	d.seen = make(map[string]struct{})
	d.mu.Unlock()
}

func fingerprint(r drift.Result) string {
	h := sha256.New()
	_, _ = fmt.Fprintf(h, "%s", r.ServiceName)
	for _, diff := range r.Diffs {
		_, _ = fmt.Fprintf(h, "|%s=%s->%s", diff.Key, diff.Declared, diff.Actual)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
