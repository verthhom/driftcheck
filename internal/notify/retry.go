package notify

import (
	"context"
	"fmt"
	"log"
	"time"

	"driftcheck/internal/drift"
)

// RetryNotifier wraps a Notifier and retries on failure up to MaxAttempts times.
type RetryNotifier struct {
	inner       drift.Notifier
	maxAttempts int
	delay       time.Duration
	logger      *log.Logger
}

// NewRetryNotifier returns a RetryNotifier that retries up to maxAttempts times
// with the given delay between attempts. A nil logger defaults to os.Stderr.
func NewRetryNotifier(inner drift.Notifier, maxAttempts int, delay time.Duration, logger *log.Logger) *RetryNotifier {
	if logger == nil {
		logger = log.Default()
	}
	if maxAttempts < 1 {
		maxAttempts = 1
	}
	return &RetryNotifier{
		inner:       inner,
		maxAttempts: maxAttempts,
		delay:       delay,
		logger:      logger,
	}
}

// Notify attempts to deliver the result, retrying on error.
func (r *RetryNotifier) Notify(ctx context.Context, result drift.Result) error {
	var lastErr error
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("retry notifier: context cancelled: %w", err)
		}
		if err := r.inner.Notify(ctx, result); err == nil {
			return nil
		} else {
			lastErr = err
			r.logger.Printf("retry notifier: attempt %d/%d failed: %v", attempt, r.maxAttempts, err)
		}
		if attempt < r.maxAttempts {
			select {
			case <-time.After(r.delay):
			case <-ctx.Done():
				return fmt.Errorf("retry notifier: context cancelled during backoff: %w", ctx.Err())
			}
		}
	}
	return fmt.Errorf("retry notifier: all %d attempts failed: %w", r.maxAttempts, lastErr)
}
