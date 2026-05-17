package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// WindowNotifier accumulates results within a rolling time window and forwards
// a combined result to the inner notifier when the window closes or Flush is
// called explicitly. Unlike BatchNotifier it replaces earlier results for the
// same service rather than appending them.
type WindowNotifier struct {
	inner    Notifier
	window   time.Duration
	mu       sync.Mutex
	latest   map[string]Result
	timer    *time.Timer
	stopOnce sync.Once
	stopCh   chan struct{}
}

// NewWindowNotifier creates a WindowNotifier that collapses repeated results
// for the same service within the given window duration before forwarding to
// inner. Panics if inner is nil or window is zero.
func NewWindowNotifier(inner Notifier, window time.Duration) *WindowNotifier {
	if inner == nil {
		panic("notify: WindowNotifier inner must not be nil")
	}
	if window <= 0 {
		panic("notify: WindowNotifier window must be positive")
	}
	w := &WindowNotifier{
		inner:  inner,
		window: window,
		latest: make(map[string]Result),
		stopCh: make(chan struct{}),
	}
	return w
}

// Notify stores the most recent result for result.Service and resets the
// flush timer. The first call within a window arms the timer.
func (w *WindowNotifier) Notify(ctx context.Context, result Result) error {
	if result.Service == "" {
		return fmt.Errorf("notify: WindowNotifier requires a non-empty service name")
	}
	w.mu.Lock()
	w.latest[result.Service] = result
	if w.timer == nil {
		w.timer = time.AfterFunc(w.window, func() { _ = w.Flush(ctx) })
	} else {
		w.timer.Reset(w.window)
	}
	w.mu.Unlock()
	return nil
}

// Flush immediately forwards all accumulated results to the inner notifier and
// clears the internal state. Returns the first error encountered.
func (w *WindowNotifier) Flush(ctx context.Context) error {
	w.mu.Lock()
	results := make([]Result, 0, len(w.latest))
	for _, r := range w.latest {
		results = append(results, r)
	}
	w.latest = make(map[string]Result)
	if w.timer != nil {
		w.timer.Stop()
		w.timer = nil
	}
	w.mu.Unlock()

	var firstErr error
	for _, r := range results {
		if err := w.inner.Notify(ctx, r); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
