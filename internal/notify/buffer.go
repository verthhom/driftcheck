package notify

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BufferedNotifier accumulates results and flushes them as a combined
// notification once the buffer reaches capacity or a flush interval elapses.
type BufferedNotifier struct {
	inner    Notifier
	capacity int
	interval time.Duration

	mu      sync.Mutex
	buf     []DriftResult
	ticker  *time.Ticker
	stopCh  chan struct{}
	done    chan struct{}
}

// NewBufferedNotifier creates a BufferedNotifier wrapping inner.
// It flushes when the buffer reaches capacity or after interval, whichever
// comes first. A zero interval disables time-based flushing.
func NewBufferedNotifier(inner Notifier, capacity int, interval time.Duration) *BufferedNotifier {
	if inner == nil {
		panic("notify: BufferedNotifier inner must not be nil")
	}
	if capacity < 1 {
		capacity = 1
	}
	b := &BufferedNotifier{
		inner:    inner,
		capacity: capacity,
		interval: interval,
		stopCh:   make(chan struct{}),
		done:     make(chan struct{}),
	}
	if interval > 0 {
		b.ticker = time.NewTicker(interval)
		go b.run()
	} else {
		close(b.done)
	}
	return b
}

// Notify buffers the result and flushes if capacity is reached.
func (b *BufferedNotifier) Notify(ctx context.Context, result DriftResult) error {
	b.mu.Lock()
	b.buf = append(b.buf, result)
	ready := len(b.buf) >= b.capacity
	var batch []DriftResult
	if ready {
		batch = b.drain()
	}
	b.mu.Unlock()
	if ready {
		return b.flush(ctx, batch)
	}
	return nil
}

// Stop halts the background flush goroutine and flushes any remaining items.
func (b *BufferedNotifier) Stop(ctx context.Context) error {
	if b.ticker != nil {
		b.ticker.Stop()
		close(b.stopCh)
		<-b.done
	}
	b.mu.Lock()
	batch := b.drain()
	b.mu.Unlock()
	if len(batch) == 0 {
		return nil
	}
	return b.flush(ctx, batch)
}

// Len returns the number of results currently buffered.
func (b *BufferedNotifier) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.buf)
}

func (b *BufferedNotifier) run() {
	defer close(b.done)
	for {
		select {
		case <-b.ticker.C:
			b.mu.Lock()
			batch := b.drain()
			b.mu.Unlock()
			if len(batch) > 0 {
				_ = b.flush(context.Background(), batch)
			}
		case <-b.stopCh:
			return
		}
	}
}

// drain returns the current buffer and resets it. Caller must hold mu.
func (b *BufferedNotifier) drain() []DriftResult {
	if len(b.buf) == 0 {
		return nil
	}
	out := make([]DriftResult, len(b.buf))
	copy(out, b.buf)
	b.buf = b.buf[:0]
	return out
}

func (b *BufferedNotifier) flush(ctx context.Context, batch []DriftResult) error {
	if len(batch) == 0 {
		return nil
	}
	combined := combine(batch)
	if err := b.inner.Notify(ctx, combined); err != nil {
		return fmt.Errorf("buffered notifier flush: %w", err)
	}
	return nil
}
