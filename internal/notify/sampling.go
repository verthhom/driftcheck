package notify

import (
	"math/rand"
	"sync"

	"driftcheck/internal/drift"
)

// SamplingNotifier forwards a random sample of notifications to the inner
// notifier. A rate of 1.0 forwards all notifications; 0.0 forwards none.
// Notifications that have drift are always forwarded regardless of the rate.
type SamplingNotifier struct {
	mu    sync.Mutex
	inner Notifier
	rate  float64
	rng   *rand.Rand
}

// NewSamplingNotifier creates a SamplingNotifier that forwards notifications
// with the given probability (0.0–1.0). Drifted results are always forwarded.
// It panics if inner is nil or rate is outside [0.0, 1.0].
func NewSamplingNotifier(inner Notifier, rate float64, seed int64) *SamplingNotifier {
	if inner == nil {
		panic("notify: SamplingNotifier inner notifier must not be nil")
	}
	if rate < 0.0 || rate > 1.0 {
		panic("notify: SamplingNotifier rate must be between 0.0 and 1.0")
	}
	return &SamplingNotifier{
		inner: inner,
		rate:  rate,
		rng:   rand.New(rand.NewSource(seed)),
	}
}

// Notify forwards the result to the inner notifier if the result has drift or
// the random sample passes the configured rate threshold.
func (s *SamplingNotifier) Notify(result drift.Result) error {
	if result.HasDrift() {
		return s.inner.Notify(result)
	}

	s.mu.Lock()
	v := s.rng.Float64()
	s.mu.Unlock()

	if v < s.rate {
		return s.inner.Notify(result)
	}
	return nil
}
