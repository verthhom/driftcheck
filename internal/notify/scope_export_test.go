package notify

// ScopeInner exposes the inner notifier for white-box testing.
func ScopeInner(n *ScopeNotifier) Notifier {
	return n.inner
}

// ScopeMap exposes the internal scopes map for white-box testing.
func ScopeMap(n *ScopeNotifier) map[string]struct{} {
	return n.scopes
}
