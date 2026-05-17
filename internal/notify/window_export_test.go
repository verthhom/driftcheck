package notify

// WindowLatestLen returns the number of pending results held in the window
// notifier. Used only in tests.
func WindowLatestLen(w *WindowNotifier) int {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(w.latest)
}

// WindowInner exposes the inner notifier for inspection in tests.
func WindowInner(w *WindowNotifier) Notifier {
	return w.inner
}
