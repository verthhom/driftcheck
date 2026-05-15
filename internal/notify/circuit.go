package notify

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"driftcheck/internal/drift"
)

// State represents the circuit breaker state.
type State int

const (
	StateClosed State = iota
	StateOpen
	StateHalfOpen
)

// ErrCircuitOpen is returned when the circuit is open and calls are rejected.
var ErrCircuitOpen = errors.New("circuit breaker is open")

// CircuitBreakerNotifier wraps a Notifier with a circuit breaker that opens
// after a threshold of consecutive failures and resets after a cooldown.
type CircuitBreakerNotifier struct {
	inner      drift.Notifier
	threshold  int
	cooldown   time.Duration

	mu          sync.Mutex
	failures    int
	state       State
	openedAt    time.Time
}

// NewCircuitBreakerNotifier creates a CircuitBreakerNotifier.
// threshold is the number of consecutive failures before opening.
// cooldown is how long to wait before moving to half-open.
func NewCircuitBreakerNotifier(inner drift.Notifier, threshold int, cooldown time.Duration) *CircuitBreakerNotifier {
	if inner == nil {
		panic("circuit breaker: inner notifier must not be nil")
	}
	if threshold <= 0 {
		threshold = 1
	}
	return &CircuitBreakerNotifier{
		inner:     inner,
		threshold: threshold,
		cooldown:  cooldown,
		state:     StateClosed,
	}
}

// Notify forwards the call to the inner notifier unless the circuit is open.
func (c *CircuitBreakerNotifier) Notify(result drift.Result) error {
	c.mu.Lock()
	switch c.state {
	case StateOpen:
		if time.Since(c.openedAt) >= c.cooldown {
			c.state = StateHalfOpen
		} else {
			c.mu.Unlock()
			return fmt.Errorf("%w: retry after %s", ErrCircuitOpen,
				c.openedAt.Add(c.cooldown).Format(time.RFC3339))
		}
	case StateClosed, StateHalfOpen:
		// proceed
	}
	c.mu.Unlock()

	err := c.inner.Notify(result)

	c.mu.Lock()
	defer c.mu.Unlock()
	if err != nil {
		c.failures++
		if c.failures >= c.threshold {
			c.state = StateOpen
			c.openedAt = time.Now()
		}
		return err
	}
	c.failures = 0
	c.state = StateClosed
	return nil
}

// CurrentState returns the current circuit state.
func (c *CircuitBreakerNotifier) CurrentState() State {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}
