package notify

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"driftcheck/internal/drift"
)

// BatchNotifier accumulates drift results over a window and flushes them
// together as a single notification to the inner Notifier.
type BatchNotifier struct {
	inner   drift.Notifier
	window  time.Duration
	mu      sync.Mutex
	buf     []drift.Result
	timer   *time.Timer
	flushFn func()
}

// NewBatchNotifier returns a BatchNotifier that collects results for the
// given window duration before forwarding a combined summary to inner.
// A zero or negative window flushes immediately (no batching).
func NewBatchNotifier(inner drift.Notifier, window time.Duration) *BatchNotifier {
	if inner == nil {
		panic("notify: BatchNotifier inner must not be nil")
	}
	bn := &BatchNotifier{
		inner:  inner,
		window: window,
	}
	bn.flushFn = bn.flush
	return bn
}

// Notify adds the result to the current batch. If the window has not yet
// started it arms a timer to flush after the configured duration.
func (b *BatchNotifier) Notify(ctx context.Context, result drift.Result) error {
	if b.window <= 0 {
		return b.inner.Notify(ctx, result)
	}

	b.mu.Lock()
	b.buf = append(b.buf, result)
	if b.timer == nil {
		b.timer = time.AfterFunc(b.window, b.flushFn)
	}
	b.mu.Unlock()
	return nil
}

// Flush forces immediate delivery of any buffered results, combining them
// into a single synthetic Result whose Diffs are the union of all buffered
// diffs. It is safe to call concurrently.
func (b *BatchNotifier) Flush(ctx context.Context) error {
	b.mu.Lock()
	results := b.buf
	b.buf = nil
	if b.timer != nil {
		b.timer.Stop()
		b.timer = nil
	}
	b.mu.Unlock()

	if len(results) == 0 {
		return nil
	}
	combined := combine(results)
	return b.inner.Notify(ctx, combined)
}

func (b *BatchNotifier) flush() {
	_ = b.Flush(context.Background())
}

func combine(results []drift.Result) drift.Result {
	if len(results) == 1 {
		return results[0]
	}
	names := make([]string, 0, len(results))
	var allDiffs []drift.Diff
	for _, r := range results {
		names = append(names, r.ServiceName)
		allDiffs = append(allDiffs, r.Diffs...)
	}
	return drift.Result{
		ServiceName: fmt.Sprintf("batch[%s]", strings.Join(names, ",")),
		Diffs:       allDiffs,
	}
}
