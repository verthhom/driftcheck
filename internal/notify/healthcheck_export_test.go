package notify

import "driftcheck/internal/drift"

// HealthCheckInner exposes the inner notifier for white-box testing.
func HealthCheckInner(h *HealthCheckNotifier) drift.Notifier {
	return h.inner
}
