package notify

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"driftcheck/internal/drift"
)

// AuditEntry records a single notification attempt.
type AuditEntry struct {
	Timestamp   time.Time
	Service     string
	HasDrift    bool
	DriftedKeys []string
	Notifier    string
	Err         error
}

// AuditNotifier wraps an inner Notifier and writes an audit log entry for
// every notification attempt, regardless of success or failure.
type AuditNotifier struct {
	inner  Notifier
	name   string
	logger *log.Logger
}

// NewAuditNotifier returns an AuditNotifier that decorates inner.
// name identifies the inner notifier in audit log lines.
// If w is nil, os.Stdout is used.
func NewAuditNotifier(inner Notifier, name string, w io.Writer) *AuditNotifier {
	if inner == nil {
		panic("audit: inner notifier must not be nil")
	}
	if w == nil {
		w = os.Stdout
	}
	return &AuditNotifier{
		inner:  inner,
		name:   name,
		logger: log.New(w, "", 0),
	}
}

// Notify forwards the result to the inner notifier and logs the outcome.
func (a *AuditNotifier) Notify(result drift.Result) error {
	err := a.inner.Notify(result)

	entry := AuditEntry{
		Timestamp:   time.Now().UTC(),
		Service:     result.ServiceName,
		HasDrift:    result.HasDrift(),
		DriftedKeys: driftedKeys(result),
		Notifier:    a.name,
		Err:         err,
	}

	status := "ok"
	if entry.Err != nil {
		status = fmt.Sprintf("error: %v", entry.Err)
	}

	a.logger.Printf(
		"audit notifier=%s service=%s drifted=%v keys=%v status=%s ts=%s",
		entry.Notifier,
		entry.Service,
		entry.HasDrift,
		entry.DriftedKeys,
		status,
		entry.Timestamp.Format(time.RFC3339),
	)

	return err
}
