package notify

import "time"

// CacheLen returns the number of entries currently held in the cache.
func CacheLen(c *CacheNotifier) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return len(c.cached)
}

// CacheEntryExpiry returns the expiry time for the named service entry.
func CacheEntryExpiry(c *CacheNotifier, service string) (time.Time, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	e, ok := c.cached[service]
	return e.expiresAt, ok
}
