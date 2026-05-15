package notify

import (
	"context"
	"fmt"

	"driftcheck/internal/drift"
)

// PriorityNotifier routes a drift result to one of two notifiers based on
// the number of drifted keys. Results with at least threshold drifted keys
// are forwarded to the high-priority notifier; all others go to the
// low-priority notifier.
type PriorityNotifier struct {
	threshold int
	high      Notifier
	low       Notifier
}

// Notifier is the shared interface expected by all notify wrappers.
type Notifier interface {
	Notify(ctx context.Context, result drift.Result) error
}

// NewPriorityNotifier creates a PriorityNotifier.
// threshold must be >= 1. high and low must not be nil.
func NewPriorityNotifier(threshold int, high, low Notifier) *PriorityNotifier {
	if threshold < 1 {
		panic("notify: PriorityNotifier threshold must be >= 1")
	}
	if high == nil {
		panic("notify: PriorityNotifier high notifier must not be nil")
	}
	if low == nil {
		panic("notify: PriorityNotifier low notifier must not be nil")
	}
	return &PriorityNotifier{threshold: threshold, high: high, low: low}
}

// Notify forwards the result to the appropriate notifier.
func (p *PriorityNotifier) Notify(ctx context.Context, result drift.Result) error {
	if len(result.Drifted) >= p.threshold {
		if err := p.high.Notify(ctx, result); err != nil {
			return fmt.Errorf("priority(high): %w", err)
		}
		return nil
	}
	if err := p.low.Notify(ctx, result); err != nil {
		return fmt.Errorf("priority(low): %w", err)
	}
	return nil
}
