package notify

import "time"

// EnvelopeNotifierInner exposes the inner Notifier for white-box testing.
func EnvelopeNotifierInner(n Notifier) Notifier {
	return n.(*envelopeNotifier).inner
}

// EnvelopeNotifierSetNow overrides the clock for deterministic tests.
func EnvelopeNotifierSetNow(n Notifier, f func() time.Time) {
	n.(*envelopeNotifier).now = f
}
