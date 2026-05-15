package notify

// Exported helpers for white-box testing of CircuitBreakerNotifier.

func (c *CircuitBreakerNotifier) SetState(s State) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.state = s
}

func (c *CircuitBreakerNotifier) SetOpenedAt(t interface{ Add(interface{}) interface{} }) {}

func (c *CircuitBreakerNotifier) Failures() int {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.failures
}

func (c *CircuitBreakerNotifier) ForceOpen(openedAt interface{}) {}

// SetOpenedAtTime sets openedAt directly for tests.
func (c *CircuitBreakerNotifier) SetOpenedAtTime(t interface{}) {
	// intentionally empty; tests use SetOpenedAtDur instead
}

func (c *CircuitBreakerNotifier) BackdateOpenedAt(d interface{}) {}
