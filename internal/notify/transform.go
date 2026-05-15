package notify

import (
	"fmt"
	"strings"

	"github.com/driftcheck/internal/drift"
)

// Transformer mutates a drift.Result before it is forwarded to a notifier.
type Transformer func(drift.Result) drift.Result

// TransformingNotifier wraps a Notifier and applies one or more Transformers
// to each Result before forwarding it.
type TransformingNotifier struct {
	inner        Notifier
	transformers []Transformer
}

// NewTransformingNotifier returns a TransformingNotifier that applies each
// transformer in order before delegating to inner.
func NewTransformingNotifier(inner Notifier, transformers ...Transformer) *TransformingNotifier {
	if inner == nil {
		panic("notify: TransformingNotifier inner must not be nil")
	}
	return &TransformingNotifier{inner: inner, transformers: transformers}
}

// Notify applies all transformers to r and then calls inner.Notify.
func (t *TransformingNotifier) Notify(r drift.Result) error {
	for _, fn := range t.transformers {
		r = fn(r)
	}
	return t.inner.Notify(r)
}

// RedactKeys returns a Transformer that replaces the values of any drifted
// keys whose names contain any of the given substrings (case-insensitive)
// with "[REDACTED]".
func RedactKeys(substrings ...string) Transformer {
	return func(r drift.Result) drift.Result {
		redacted := make([]drift.Difference, len(r.Diffs))
		for i, d := range r.Diffs {
			for _, sub := range substrings {
				if strings.Contains(strings.ToLower(d.Key), strings.ToLower(sub)) {
					d.Got = "[REDACTED]"
					d.Want = "[REDACTED]"
					break
				}
			}
			redacted[i] = d
		}
		r.Diffs = redacted
		return r
	}
}

// PrefixService returns a Transformer that prepends prefix to the service name.
func PrefixService(prefix string) Transformer {
	return func(r drift.Result) drift.Result {
		r.Service = fmt.Sprintf("%s%s", prefix, r.Service)
		return r
	}
}
