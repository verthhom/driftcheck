package notify

import (
	"fmt"
	"sync"
	"time"

	"driftcheck/internal/drift"
)

// RateLimitedNotifier wraps a Notifier and suppresses duplicate alerts
// for the same service within a configurable cooldown window.
type RateLimitedNotifier struct {
	inner    drift.Notifier
	cooldown time.Duration

	mu   sync.Mutex
	last map[string]time.Time
}

// NewRateLimitedNotifier returns a Notifier that forwards to inner at most
// once per cooldown duration for each service.
func NewRateLimitedNotifier(inner drift.Notifier, cooldown time.Duration) (*RateLimitedNotifier, error) {
	if inner == nil {
		return nil, fmt.Errorf("ratelimit: inner notifier must not be nil")
	}
	if cooldown <= 0 {
		return nil, fmt.Errorf("ratelimit: cooldown must be positive")
	}
	return &RateLimitedNotifier{
		inner:    inner,
		cooldown: cooldown,
		last:     make(map[string]time.Time),
	}, nil
}

// Notify forwards the result to the inner notifier only if the cooldown
// period has elapsed since the last notification for the same service.
func (r *RateLimitedNotifier) Notify(result drift.Result) error {
	service := result.ServiceName

	r.mu.Lock()
	defer r.mu.Unlock()

	if t, ok := r.last[service]; ok {
		if time.Since(t) < r.cooldown {
			return nil // suppressed
		}
	}

	if err := r.inner.Notify(result); err != nil {
		return err
	}

	r.last[service] = time.Now()
	return nil
}

// Reset clears the rate-limit state for a specific service, allowing the
// next notification to pass through immediately.
func (r *RateLimitedNotifier) Reset(service string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.last, service)
}
