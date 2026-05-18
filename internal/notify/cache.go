package notify

import (
	"fmt"
	"sync"
	"time"

	"driftcheck/internal/drift"
)

// CacheNotifier wraps an inner Notifier and skips sending if the result for a
// given service has not changed since the last successful notification.
type CacheNotifier struct {
	inner   Notifier
	ttl     time.Duration
	mu      sync.Mutex
	cached  map[string]cacheEntry
}

type cacheEntry struct {
	fingerprint string
	expiresAt   time.Time
}

// NewCacheNotifier returns a CacheNotifier that suppresses repeat notifications
// for the same result within ttl. Panics if inner is nil or ttl is zero.
func NewCacheNotifier(inner Notifier, ttl time.Duration) *CacheNotifier {
	if inner == nil {
		panic("notify: CacheNotifier inner must not be nil")
	}
	if ttl <= 0 {
		panic("notify: CacheNotifier ttl must be positive")
	}
	return &CacheNotifier{
		inner:  inner,
		ttl:    ttl,
		cached: make(map[string]cacheEntry),
	}
}

// Notify forwards the result to the inner notifier only when the fingerprint
// differs from the cached value or the TTL has expired.
func (c *CacheNotifier) Notify(result drift.Result) error {
	key := result.ServiceName
	fp := cacheFingerprint(result)

	c.mu.Lock()
	entry, ok := c.cached[key]
	now := time.Now()
	if ok && entry.fingerprint == fp && now.Before(entry.expiresAt) {
		c.mu.Unlock()
		return nil
	}
	c.cached[key] = cacheEntry{fingerprint: fp, expiresAt: now.Add(c.ttl)}
	c.mu.Unlock()

	return c.inner.Notify(result)
}

// Invalidate removes the cached entry for the given service, forcing the next
// notification to be forwarded regardless of TTL.
func (c *CacheNotifier) Invalidate(service string) {
	c.mu.Lock()
	delete(c.cached, service)
	c.mu.Unlock()
}

func cacheFingerprint(r drift.Result) string {
	return fmt.Sprintf("%s|%v|%d", r.ServiceName, r.HasDrift(), len(r.Drifted))
}
