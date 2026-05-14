package notify

import (
	"fmt"
	"sync"
	"time"

	"driftcheck/internal/drift"
)

// ThrottleNotifier wraps a Notifier and ensures that at most one notification
// is sent per service within a given window, regardless of how many drift
// checks fire during that period.
type ThrottleNotifier struct {
	inner  drift.Notifier
	window time.Duration
	mu     sync.Mutex
	last   map[string]time.Time
	now    func() time.Time
}

// NewThrottleNotifier returns a ThrottleNotifier that forwards to inner at most
// once per window for each unique service name.
// A zero or negative window disables throttling (every call is forwarded).
func NewThrottleNotifier(inner drift.Notifier, window time.Duration) *ThrottleNotifier {
	if inner == nil {
		panic("throttle: inner notifier must not be nil")
	}
	return &ThrottleNotifier{
		inner:  inner,
		window: window,
		last:   make(map[string]time.Time),
		now:    time.Now,
	}
}

// Notify forwards the result to the inner notifier only when the throttle
// window for the result's service has elapsed since the last forwarded call.
func (t *ThrottleNotifier) Notify(result drift.Result) error {
	if t.window <= 0 {
		return t.inner.Notify(result)
	}

	service := result.ServiceName
	if service == "" {
		service = "__unknown__"
	}

	t.mu.Lock()
	now := t.now()
	if last, ok := t.last[service]; ok && now.Sub(last) < t.window {
		t.mu.Unlock()
		return fmt.Errorf("throttle: notification suppressed for %q (window %s)", service, t.window)
	}
	t.last[service] = now
	t.mu.Unlock()

	return t.inner.Notify(result)
}
