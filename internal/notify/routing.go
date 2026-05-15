package notify

import (
	"context"
	"fmt"

	"driftcheck/internal/drift"
)

// RouteRule maps a predicate to a target Notifier.
type RouteRule struct {
	Match    func(drift.Result) bool
	Notifier Notifier
}

// RoutingNotifier dispatches a Result to the first matching rule.
// If no rule matches and a fallback is set, the fallback is used.
// If no rule matches and there is no fallback, Notify is a no-op.
type RoutingNotifier struct {
	rules    []RouteRule
	fallback Notifier
}

// NewRoutingNotifier creates a RoutingNotifier with the given rules.
// Rules are evaluated in order; the first match wins.
// fallback may be nil.
func NewRoutingNotifier(fallback Notifier, rules ...RouteRule) *RoutingNotifier {
	if len(rules) == 0 {
		panic("notify: RoutingNotifier requires at least one rule")
	}
	return &RoutingNotifier{rules: rules, fallback: fallback}
}

// Notify evaluates each rule in order and forwards to the first match.
func (r *RoutingNotifier) Notify(ctx context.Context, result drift.Result) error {
	for _, rule := range r.rules {
		if rule.Match(result) {
			if rule.Notifier == nil {
				return fmt.Errorf("notify: matched rule has nil notifier")
			}
			return rule.Notifier.Notify(ctx, result)
		}
	}
	if r.fallback != nil {
		return r.fallback.Notify(ctx, result)
	}
	return nil
}
