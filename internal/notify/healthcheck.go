package notify

import (
	"context"
	"fmt"
	"sync"
	"time"

	"driftcheck/internal/drift"
)

// HealthStatus represents the current health of a notifier.
type HealthStatus struct {
	Healthy      bool
	LastSuccess  time.Time
	LastFailure  time.Time
	FailureCount int
	Message      string
}

// HealthCheckNotifier wraps a Notifier and tracks its health over time.
type HealthCheckNotifier struct {
	inner   drift.Notifier
	mu      sync.RWMutex
	status  HealthStatus
}

// NewHealthCheckNotifier returns a HealthCheckNotifier wrapping inner.
// It panics if inner is nil.
func NewHealthCheckNotifier(inner drift.Notifier) *HealthCheckNotifier {
	if inner == nil {
		panic("healthcheck: inner notifier must not be nil")
	}
	return &HealthCheckNotifier{
		inner: inner,
		status: HealthStatus{Healthy: true},
	}
}

// Notify forwards the result to the inner notifier and records the outcome.
func (h *HealthCheckNotifier) Notify(ctx context.Context, result drift.Result) error {
	err := h.inner.Notify(ctx, result)

	h.mu.Lock()
	defer h.mu.Unlock()

	if err != nil {
		h.status.Healthy = false
		h.status.LastFailure = time.Now()
		h.status.FailureCount++
		h.status.Message = fmt.Sprintf("last error: %v", err)
	} else {
		h.status.Healthy = true
		h.status.LastSuccess = time.Now()
		h.status.FailureCount = 0
		h.status.Message = ""
	}

	return err
}

// Health returns a snapshot of the current health status.
func (h *HealthCheckNotifier) Health() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.status
}

// Reset clears the failure count and marks the notifier healthy.
func (h *HealthCheckNotifier) Reset() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.status.Healthy = true
	h.status.FailureCount = 0
	h.status.Message = ""
}
