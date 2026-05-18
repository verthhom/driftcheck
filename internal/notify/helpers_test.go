package notify_test

import (
	"driftcheck/internal/drift"
)

// captureNotifier records how many times Notify was called.
type captureNotifier struct {
	count  int
	last   drift.Result
}

func (c *captureNotifier) Notify(r drift.Result) error {
	c.count++
	c.last = r
	return nil
}

// errorNotifier always returns the configured error.
type errorNotifier struct {
	count int
	err   error
}

func (e *errorNotifier) Notify(_ drift.Result) error {
	e.count++
	return e.err
}
