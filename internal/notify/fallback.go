package notify

import (
	"context"
	"fmt"
	"log"
	"os"

	"driftcheck/internal/drift"
)

// FallbackNotifier attempts the primary notifier and, on failure, delegates
// to one or more fallback notifiers in order. The first successful send wins.
type FallbackNotifier struct {
	primary   drift.Notifier
	fallbacks []drift.Notifier
	logger    *log.Logger
}

// NewFallbackNotifier returns a FallbackNotifier that tries primary first,
// then each fallback in order until one succeeds.
// Panics if primary is nil or no fallbacks are provided.
func NewFallbackNotifier(primary drift.Notifier, fallbacks ...drift.Notifier) *FallbackNotifier {
	if primary == nil {
		panic("notify: FallbackNotifier primary must not be nil")
	}
	if len(fallbacks) == 0 {
		panic("notify: FallbackNotifier requires at least one fallback")
	}
	return &FallbackNotifier{
		primary:   primary,
		fallbacks: fallbacks,
		logger:    log.New(os.Stderr, "[fallback] ", log.LstdFlags),
	}
}

// WithLogger replaces the default stderr logger.
func (f *FallbackNotifier) WithLogger(l *log.Logger) *FallbackNotifier {
	f.logger = l
	return f
}

// Notify attempts the primary notifier. If it returns an error, each fallback
// is tried in order. Returns an error only when all notifiers fail.
func (f *FallbackNotifier) Notify(ctx context.Context, result drift.Result) error {
	if err := f.primary.Notify(ctx, result); err == nil {
		return nil
	} else {
		f.logger.Printf("primary notifier failed: %v; trying fallbacks", err)
	}

	var lastErr error
	for i, fb := range f.fallbacks {
		if err := fb.Notify(ctx, result); err == nil {
			f.logger.Printf("fallback[%d] succeeded", i)
			return nil
		} else {
			f.logger.Printf("fallback[%d] failed: %v", i, err)
			lastErr = err
		}
	}

	return fmt.Errorf("all notifiers failed; last error: %w", lastErr)
}
