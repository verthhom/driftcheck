package notify

import (
	"fmt"
	"io"
	"os"
	"time"

	"driftcheck/internal/drift"
)

// LogNotifier writes a structured log line for every drift result it receives.
// It is useful as a lightweight sink when a full alerting pipeline is not needed.
type LogNotifier struct {
	w      io.Writer
	prefix string
}

// LogOption configures a LogNotifier.
type LogOption func(*LogNotifier)

// WithLogWriter redirects log output to w instead of os.Stdout.
func WithLogWriter(w io.Writer) LogOption {
	return func(n *LogNotifier) { n.w = w }
}

// WithLogPrefix prepends a static label to every log line.
func WithLogPrefix(prefix string) LogOption {
	return func(n *LogNotifier) { n.prefix = prefix }
}

// NewLogNotifier returns a LogNotifier with the supplied options applied.
func NewLogNotifier(opts ...LogOption) *LogNotifier {
	n := &LogNotifier{w: os.Stdout}
	for _, o := range opts {
		o(n)
	}
	return n
}

// Notify writes one log line per result to the configured writer.
func (n *LogNotifier) Notify(results []drift.Result) error {
	ts := time.Now().UTC().Format(time.RFC3339)
	for _, r := range results {
		status := "ok"
		if r.HasDrift() {
			status = "drift"
		}
		line := fmt.Sprintf("%s [%s] service=%s drifted_keys=%d status=%s\n",
			ts, n.prefix, r.ServiceName, len(r.Diffs), status)
		if _, err := fmt.Fprint(n.w, line); err != nil {
			return fmt.Errorf("log notifier write: %w", err)
		}
	}
	return nil
}
