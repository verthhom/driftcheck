package drift

import (
	"fmt"
	"io"
	"strings"
)

// Reporter formats and writes DriftResult output.
type Reporter struct {
	Writer io.Writer
}

// NewReporter creates a Reporter that writes to the given writer.
func NewReporter(w io.Writer) *Reporter {
	return &Reporter{Writer: w}
}

// Report writes a human-readable summary of the drift result.
func (r *Reporter) Report(result DriftResult) {
	if !result.Drifted {
		fmt.Fprintf(r.Writer, "[OK] %s: no drift detected\n", result.ServiceName)
		return
	}

	fmt.Fprintf(r.Writer, "[DRIFT] %s: %d difference(s) found\n", result.ServiceName, len(result.Diffs))
	for _, d := range result.Diffs {
		fmt.Fprintf(r.Writer, "  key=%q declared=%q deployed=%q\n", d.Key, d.Declared, d.Deployed)
	}
}

// ReportAll writes a summary for multiple drift results and returns an exit-worthy status.
func (r *Reporter) ReportAll(results []DriftResult) bool {
	anyDrift := false
	for _, res := range results {
		r.Report(res)
		if res.Drifted {
			anyDrift = true
		}
	}

	fmt.Fprintln(r.Writer, strings.Repeat("-", 40))
	if anyDrift {
		fmt.Fprintln(r.Writer, "Result: DRIFT DETECTED")
	} else {
		fmt.Fprintln(r.Writer, "Result: ALL SERVICES IN SYNC")
	}
	return anyDrift
}
