package notify

import (
	"errors"
	"sync"

	"github.com/driftcheck/internal/drift"
)

// SpoolNotifier buffers results in memory when the inner notifier is
// unavailable and drains them in FIFO order on the next successful Notify
// call. Unlike BufferedNotifier (which batches by capacity), SpoolNotifier
// is designed for fault-tolerance: it keeps trying to drain on every call.
type SpoolNotifier struct {
	mu      sync.Mutex
	inner   Notifier
	spool   []drift.Result
	maxSize int
}

// NewSpoolNotifier creates a SpoolNotifier wrapping inner.
// maxSize is the maximum number of results held in the spool; once full,
// the oldest entry is dropped to make room (head-drop policy).
func NewSpoolNotifier(inner Notifier, maxSize int) *SpoolNotifier {
	if inner == nil {
		panic("notify: SpoolNotifier inner must not be nil")
	}
	if maxSize <= 0 {
		panic("notify: SpoolNotifier maxSize must be positive")
	}
	return &SpoolNotifier{inner: inner, maxSize: maxSize}
}

// Notify attempts to drain any spooled results before forwarding r.
// If the inner notifier fails, r is appended to the spool for later retry.
func (s *SpoolNotifier) Notify(r drift.Result) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Drain existing spool first.
	remaining := s.spool[:0]
	for _, queued := range s.spool {
		if err := s.inner.Notify(queued); err != nil {
			remaining = append(remaining, queued)
		}
	}
	s.spool = remaining

	// Forward the new result.
	if err := s.inner.Notify(r); err != nil {
		s.enqueue(r)
		return err
	}
	return nil
}

// Len returns the number of results currently held in the spool.
func (s *SpoolNotifier) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.spool)
}

// Drain discards all spooled results and returns them to the caller.
func (s *SpoolNotifier) Drain() []drift.Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]drift.Result, len(s.spool))
	copy(out, s.spool)
	s.spool = nil
	return out
}

var ErrSpoolFull = errors.New("notify: spool is full, oldest entry dropped")

func (s *SpoolNotifier) enqueue(r drift.Result) {
	if len(s.spool) >= s.maxSize {
		// Head-drop: remove oldest.
		s.spool = s.spool[1:]
	}
	s.spool = append(s.spool, r)
}
