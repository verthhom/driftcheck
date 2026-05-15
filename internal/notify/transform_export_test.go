package notify

// TransformingNotifierInner exposes the inner notifier for white-box tests.
func TransformingNotifierInner(t *TransformingNotifier) Notifier {
	return t.inner
}

// TransformingNotifierCount returns the number of transformers registered.
func TransformingNotifierCount(t *TransformingNotifier) int {
	return len(t.transformers)
}
