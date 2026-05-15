package notify

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"driftcheck/internal/drift"
)

// DigestNotifier accumulates drift results over a window and emits a
// single summarised notification when the window closes or Flush is
// called explicitly.
type DigestNotifier struct {
	mu       sync.Mutex
	inner    Notifier
	window   time.Duration
	results  []drift.Result
	timer    *time.Timer
	out      io.Writer
}

// NewDigestNotifier returns a DigestNotifier that batches results for
// the given window duration before forwarding a digest to inner.
// A zero window disables automatic flushing; callers must call Flush.
func NewDigestNotifier(inner Notifier, window time.Duration) *DigestNotifier {
	if inner == nil {
		panic("digest: inner notifier must not be nil")
	}
	d := &DigestNotifier{
		inner:  inner,
		window: window,
		out:    os.Stdout,
	}
	if window > 0 {
		d.timer = time.AfterFunc(window, func() { _ = d.Flush() })
	}
	return d
}

// Notify queues the result for the next digest.
func (d *DigestNotifier) Notify(r drift.Result) error {
	d.mu.Lock()
	d.results = append(d.results, r)
	d.mu.Unlock()
	return nil
}

// Flush emits a single combined notification for all queued results
// and resets the internal queue. It restarts the window timer if one
// was configured.
func (d *DigestNotifier) Flush() error {
	d.mu.Lock()
	if len(d.results) == 0 {
		d.mu.Unlock()
		return nil
	}
	snap := make([]drift.Result, len(d.results))
	copy(snap, d.results)
	d.results = d.results[:0]
	d.mu.Unlock()

	combined := mergeResults(snap)

	if d.window > 0 && d.timer != nil {
		d.timer.Reset(d.window)
	}
	return d.inner.Notify(combined)
}

// mergeResults combines multiple drift.Results into one digest result.
// The service name is a comma-separated list of unique names; drifted
// keys are the union of all individual drifted keys.
func mergeResults(results []drift.Result) drift.Result {
	seen := make(map[string]struct{})
	var services []string
	var allDrifted []drift.KeyDiff

	for _, r := range results {
		if _, ok := seen[r.ServiceName]; !ok {
			seen[r.ServiceName] = struct{}{}
			services = append(services, r.ServiceName)
		}
		allDrifted = append(allDrifted, r.Drifted...)
	}

	return drift.Result{
		ServiceName: fmt.Sprintf("digest[%s]", strings.Join(services, ",")),
		Drifted:     allDrifted,
	}
}
