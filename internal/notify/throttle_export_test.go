package notify

import "time"

// SetNow replaces the clock used by ThrottleNotifier for testing.
func (t *ThrottleNotifier) SetNow(fn func() time.Time) {
	t.now = fn
}
