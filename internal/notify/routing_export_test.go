package notify

// RoutingRules exposes the rules slice for white-box testing.
func RoutingRules(n *RoutingNotifier) []RouteRule { return n.rules }

// RoutingFallback exposes the fallback notifier for white-box testing.
func RoutingFallback(n *RoutingNotifier) Notifier { return n.fallback }
