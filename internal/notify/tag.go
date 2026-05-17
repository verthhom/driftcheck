package notify

import (
	"context"
	"fmt"
	"strings"

	"driftcheck/internal/drift"
)

// TaggingNotifier wraps an inner Notifier and attaches a fixed set of key/value
// tags to every result's service name, making it easy to distinguish results
// produced by a particular environment, region, or deployment tier.
type TaggingNotifier struct {
	inner  Notifier
	tags   map[string]string
	prefix string // computed once from tags
}

// NewTaggingNotifier returns a TaggingNotifier that forwards every call to
// inner after annotating the service name with the supplied tags.
//
// Tags are appended to the service name in the form:
//
//	<service>[key=value,key=value]
//
// Panics if inner is nil or tags is empty.
func NewTaggingNotifier(inner Notifier, tags map[string]string) *TaggingNotifier {
	if inner == nil {
		panic("notify: NewTaggingNotifier: inner must not be nil")
	}
	if len(tags) == 0 {
		panic("notify: NewTaggingNotifier: tags must not be empty")
	}

	parts := make([]string, 0, len(tags))
	for k, v := range tags {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	prefix := "[" + strings.Join(parts, ",") + "]"

	return &TaggingNotifier{inner: inner, tags: tags, prefix: prefix}
}

// Notify annotates result.ServiceName with the configured tags then forwards
// the modified result to the inner Notifier.
func (t *TaggingNotifier) Notify(ctx context.Context, result drift.Result) error {
	annotated := result
	annotated.ServiceName = result.ServiceName + t.prefix
	return t.inner.Notify(ctx, annotated)
}
