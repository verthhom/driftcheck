package notify

import (
	"fmt"
	"time"
)

// Envelope wraps a notification result with metadata before forwarding it
// to an inner Notifier. This is useful for adding consistent context such as
// environment tags, host identifiers, or schema versions to every alert.
type Envelope struct {
	Environment string
	Host        string
	SchemaVersion string
}

// envelopeNotifier applies an Envelope to each result before delegating.
type envelopeNotifier struct {
	inner    Notifier
	envelope Envelope
	now      func() time.Time
}

// NewEnvelopeNotifier returns a Notifier that stamps each result with the
// provided Envelope metadata. It panics if inner is nil.
func NewEnvelopeNotifier(inner Notifier, env Envelope) Notifier {
	if inner == nil {
		panic("notify: NewEnvelopeNotifier: inner must not be nil")
	}
	return &envelopeNotifier{
		inner:    inner,
		envelope: env,
		now:      time.Now,
	}
}

// Notify stamps r with envelope metadata and forwards it to the inner Notifier.
func (n *envelopeNotifier) Notify(r Result) error {
	stamped := r
	if stamped.Meta == nil {
		stamped.Meta = make(map[string]string)
	}
	if n.envelope.Environment != "" {
		stamped.Meta["environment"] = n.envelope.Environment
	}
	if n.envelope.Host != "" {
		stamped.Meta["host"] = n.envelope.Host
	}
	if n.envelope.SchemaVersion != "" {
		stamped.Meta["schema_version"] = n.envelope.SchemaVersion
	}
	stamped.Meta["envelope_ts"] = fmt.Sprintf("%d", n.now().Unix())
	return n.inner.Notify(stamped)
}
