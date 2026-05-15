package notify

// DigestLen returns the number of queued results inside d.
// Exported for white-box testing only.
func DigestLen(d *DigestNotifier) int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.results)
}

// DigestInner returns the inner notifier.
func DigestInner(d *DigestNotifier) Notifier {
	return d.inner
}
