package notify

import (
	"context"
	"fmt"

	"github.com/driftcheck/driftcheck/internal/drift"
)

// ScopeNotifier forwards notifications only when the result's service name
// matches one of the allowed scopes. It is useful when a single notifier
// pipeline is shared across many services but certain downstream targets
// should only receive events for a defined subset.
type ScopeNotifier struct {
	inner  Notifier
	scopes map[string]struct{}
}

// NewScopeNotifier returns a ScopeNotifier that forwards to inner only when
// the result belongs to one of the provided service names.
// Panics if inner is nil or no scopes are provided.
func NewScopeNotifier(inner Notifier, services ...string) *ScopeNotifier {
	if inner == nil {
		panic("scope: inner notifier must not be nil")
	}
	if len(services) == 0 {
		panic("scope: at least one service name is required")
	}
	scopes := make(map[string]struct{}, len(services))
	for _, s := range services {
		scopes[s] = struct{}{}
	}
	return &ScopeNotifier{inner: inner, scopes: scopes}
}

// Notify forwards r to the inner notifier only when r.Service is in scope.
// Results for out-of-scope services are silently dropped and nil is returned.
func (n *ScopeNotifier) Notify(ctx context.Context, r drift.Result) error {
	if _, ok := n.scopes[r.Service]; !ok {
		return nil
	}
	return n.inner.Notify(ctx, r)
}

// Scopes returns the set of service names this notifier accepts.
func (n *ScopeNotifier) Scopes() []string {
	out := make([]string, 0, len(n.scopes))
	for s := range n.scopes {
		out = append(out, s)
	}
	return out
}

// String implements fmt.Stringer for logging purposes.
func (n *ScopeNotifier) String() string {
	return fmt.Sprintf("ScopeNotifier(scopes=%d)", len(n.scopes))
}
