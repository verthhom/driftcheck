package notify

import (
	"context"

	"github.com/example/driftcheck/internal/drift"
)

// FilterFunc is a predicate that decides whether a notification should be sent
// for the given drift result. Returning false suppresses the notification.
type FilterFunc func(result drift.Result) bool

// FilteredNotifier wraps a Notifier and only forwards calls that satisfy all
// registered filter predicates.
type FilteredNotifier struct {
	inner   Notifier
	filters []FilterFunc
}

// NewFilteredNotifier returns a FilteredNotifier that delegates to inner only
// when every filter returns true. Passing no filters means every result is
// forwarded.
func NewFilteredNotifier(inner Notifier, filters ...FilterFunc) (*FilteredNotifier, error) {
	if inner == nil {
		return nil, ErrNilInner
	}
	return &FilteredNotifier{inner: inner, filters: filters}, nil
}

// Notify forwards the result to the inner notifier only when all filters pass.
func (f *FilteredNotifier) Notify(ctx context.Context, result drift.Result) error {
	for _, fn := range f.filters {
		if !fn(result) {
			return nil
		}
	}
	return f.inner.Notify(ctx, result)
}

// OnlyDrifted is a FilterFunc that suppresses notifications when there is no
// detected drift.
func OnlyDrifted(result drift.Result) bool {
	return result.HasDrift()
}

// OnlyService returns a FilterFunc that passes results only for the named
// service.
func OnlyService(name string) FilterFunc {
	return func(result drift.Result) bool {
		return result.ServiceName == name
	}
}
