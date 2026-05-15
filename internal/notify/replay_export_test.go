package notify

// ReplayBuf exposes the internal buffer length for white-box tests.
func ReplayBuf(r *ReplayNotifier) int { return r.Len() }

// ReplayInner exposes the wrapped notifier for white-box tests.
func ReplayInner(r *ReplayNotifier) Notifier { return r.inner }
