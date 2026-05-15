package notify

import (
	"context"
	"fmt"
	"time"

	"driftcheck/internal/drift"
)

// TimeoutNotifier wraps an inner Notifier and cancels the Notify call if it
// exceeds the configured deadline. If the inner notifier does not respect
// context cancellation the goroutine may linger, but the caller will receive
// an error promptly.
type TimeoutNotifier struct {
	inner   Notifier
	timeout time.Duration
}

// NewTimeoutNotifier returns a TimeoutNotifier that enforces d as a maximum
// wall-clock budget for every Notify call. It panics when inner is nil or d
// is not positive.
func NewTimeoutNotifier(inner Notifier, d time.Duration) *TimeoutNotifier {
	if inner == nil {
		panic("notify: TimeoutNotifier requires a non-nil inner notifier")
	}
	if d <= 0 {
		panic("notify: TimeoutNotifier requires a positive timeout")
	}
	return &TimeoutNotifier{inner: inner, timeout: d}
}

// Notify forwards result to the inner notifier, returning an error if the
// deadline is exceeded before the inner call returns.
func (t *TimeoutNotifier) Notify(ctx context.Context, result drift.Result) error {
	ctx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()

	type outcome struct {
		err error
	}
	ch := make(chan outcome, 1)

	go func() {
		ch <- outcome{err: t.inner.Notify(ctx, result)}
	}()

	select {
	case out := <-ch:
		return out.err
	case <-ctx.Done():
		return fmt.Errorf("notify: timeout after %s: %w", t.timeout, ctx.Err())
	}
}
