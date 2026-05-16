package notify

import (
	"fmt"
	"sync"

	"github.com/driftcheck/internal/drift"
)

// SquashNotifier collapses multiple Notify calls for the same service within a
// flush window into a single call, keeping only the most-recent result. Unlike
// BatchNotifier (which combines all results), SquashNotifier discards older
// results so the inner notifier always sees the latest state.
type SquashNotifier struct {
	mu      sync.Mutex
	inner   Notifier
	pending map[string]drift.Result
}

// NewSquashNotifier returns a SquashNotifier wrapping inner.
// Callers must call Flush to forward accumulated results.
func NewSquashNotifier(inner Notifier) *SquashNotifier {
	if inner == nil {
		panic("squash: inner notifier must not be nil")
	}
	return &SquashNotifier{
		inner:   inner,
		pending: make(map[string]drift.Result),
	}
}

// Notify stores result, overwriting any previously queued result for the same
// service name. It never calls the inner notifier directly.
func (s *SquashNotifier) Notify(result drift.Result) error {
	if result.ServiceName == "" {
		return fmt.Errorf("squash: result has empty service name")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pending[result.ServiceName] = result
	return nil
}

// Flush forwards the latest result for every queued service to the inner
// notifier and clears the queue. All inner errors are collected and returned
// as a single combined error.
func (s *SquashNotifier) Flush() error {
	s.mu.Lock()
	snap := s.pending
	s.pending = make(map[string]drift.Result)
	s.mu.Unlock()

	var errs []error
	for _, result := range snap {
		if err := s.inner.Notify(result); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("squash: %d flush error(s): %v", len(errs), errs)
}

// Len returns the number of services currently queued.
func (s *SquashNotifier) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.pending)
}
