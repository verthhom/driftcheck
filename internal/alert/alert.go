// Package alert provides notification mechanisms for detected configuration drift.
package alert

import (
	"fmt"
	"io"
	"os"
	"time"
)

// Severity represents the urgency level of a drift alert.
type Severity string

const (
	SeverityInfo     Severity = "INFO"
	SeverityWarning  Severity = "WARNING"
	SeverityCritical Severity = "CRITICAL"
)

// Alert represents a notification about detected drift.
type Alert struct {
	ServiceName string
	Severity    Severity
	Message     string
	DriftKeys   []string
	OccurredAt  time.Time
}

// Notifier sends alerts to a destination.
type Notifier interface {
	Notify(a Alert) error
}

// LogNotifier writes alerts to an io.Writer.
type LogNotifier struct {
	Writer io.Writer
}

// NewLogNotifier creates a LogNotifier that writes to w.
// If w is nil, os.Stdout is used.
func NewLogNotifier(w io.Writer) *LogNotifier {
	if w == nil {
		w = os.Stdout
	}
	return &LogNotifier{Writer: w}
}

// Notify formats and writes the alert to the underlying writer.
func (n *LogNotifier) Notify(a Alert) error {
	_, err := fmt.Fprintf(
		n.Writer,
		"[%s] %s | service=%s drifted_keys=%v\n",
		a.Severity,
		a.OccurredAt.UTC().Format(time.RFC3339),
		a.ServiceName,
		a.DriftKeys,
	)
	return err
}

// SeverityFor returns the appropriate Severity based on the number of drifted keys.
func SeverityFor(driftCount int) Severity {
	switch {
	case driftCount == 0:
		return SeverityInfo
	case driftCount <= 3:
		return SeverityWarning
	default:
		return SeverityCritical
	}
}
