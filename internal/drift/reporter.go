package drift

import (
	"fmt"
	"io"
	"os"
	"text/tabwriter"
	"time"
)

// Result holds the outcome of a single drift check.
type Result struct {
	ServiceName string
	DriftedKeys []DriftedKey
	CheckedAt   time.Time
}

// DriftedKey describes a single key that has drifted.
type DriftedKey struct {
	Key      string
	Expected string
	Actual   string
	Reason   string
}

// HasDrift returns true when at least one key has drifted.
func (r Result) HasDrift() bool {
	return len(r.DriftedKeys) > 0
}

// Reporter formats and writes drift results.
type Reporter struct {
	out io.Writer
}

// NewReporter creates a Reporter that writes to out.
// If out is nil, os.Stdout is used.
func NewReporter(out io.Writer) *Reporter {
	if out == nil {
		out = os.Stdout
	}
	return &Reporter{out: out}
}

// Report writes a human-readable summary of the drift result.
func (r *Reporter) Report(result Result) {
	if !result.HasDrift() {
		fmt.Fprintf(r.out, "[OK] %s — no drift detected (checked at %s)\n",
			result.ServiceName, result.CheckedAt.Format(time.RFC3339))
		return
	}

	fmt.Fprintf(r.out, "[DRIFT] %s — %d key(s) drifted (checked at %s)\n",
		result.ServiceName, len(result.DriftedKeys), result.CheckedAt.Format(time.RFC3339))

	tw := tabwriter.NewWriter(r.out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "  KEY\tEXPECTED\tACTUAL\tREASON")
	for _, dk := range result.DriftedKeys {
		fmt.Fprintf(tw, "  %s\t%s\t%s\t%s\n", dk.Key, dk.Expected, dk.Actual, dk.Reason)
	}
	_ = tw.Flush()
}

// ReportAll writes a summary for each result in the provided slice and returns
// the total number of results that contained drift.
func (r *Reporter) ReportAll(results []Result) int {
	driftCount := 0
	for _, result := range results {
		r.Report(result)
		if result.HasDrift() {
			driftCount++
		}
	}
	return driftCount
}
