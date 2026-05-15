package notify

import "time"

// SetOpenedAtDur backdates the openedAt timestamp by the given duration,
// allowing tests to simulate elapsed cooldown time.
func (c *CircuitBreakerNotifier) SetOpenedAtDur(ago time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.openedAt = time.Now().Add(-ago)
}

// ResetFailures resets the failure counter without changing state.
func (c *CircuitBreakerNotifier) ResetFailures() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.failures = 0
}
