package notify

import (
	"context"
	"fmt"
	"sync"

	"driftcheck/internal/drift"
)

// ReplayNotifier stores a fixed number of recent results and can replay
// them to a target notifier on demand. Useful for bootstrapping new
// notification sinks with recent history.
type ReplayNotifier struct {
	mu       sync.Mutex
	inner    Notifier
	capacity int
	buf      []drift.Result
}

// NewReplayNotifier creates a ReplayNotifier that retains up to capacity
// results. Panics if inner is nil or capacity is less than 1.
func NewReplayNotifier(inner Notifier, capacity int) *ReplayNotifier {
	if inner == nil {
		panic("notify: ReplayNotifier inner notifier must not be nil")
	}
	if capacity < 1 {
		panic("notify: ReplayNotifier capacity must be at least 1")
	}
	return &ReplayNotifier{
		inner:    inner,
		capacity: capacity,
		buf:      make([]drift.Result, 0, capacity),
	}
}

// Notify forwards the result to the inner notifier and records it in the
// replay buffer. When the buffer is full the oldest entry is evicted.
func (r *ReplayNotifier) Notify(ctx context.Context, result drift.Result) error {
	r.mu.Lock()
	if len(r.buf) >= r.capacity {
		r.buf = r.buf[1:]
	}
	r.buf = append(r.buf, result)
	r.mu.Unlock()

	return r.inner.Notify(ctx, result)
}

// Replay sends all buffered results to target in the order they were
// received. Errors are collected and returned as a combined error; replay
// continues even when individual sends fail.
func (r *ReplayNotifier) Replay(ctx context.Context, target Notifier) error {
	r.mu.Lock()
	snap := make([]drift.Result, len(r.buf))
	copy(snap, r.buf)
	r.mu.Unlock()

	var errs []error
	for _, res := range snap {
		if err := target.Notify(ctx, res); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("notify: replay encountered %d error(s): %w", len(errs), errs[0])
}

// Len returns the number of results currently held in the buffer.
func (r *ReplayNotifier) Len() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.buf)
}
