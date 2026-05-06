package history

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
	"time"
)

// Exporter writes drift history records to an output format.
type Exporter struct {
	w io.Writer
}

// NewExporter creates an Exporter that writes to w.
func NewExporter(w io.Writer) *Exporter {
	return &Exporter{w: w}
}

// ExportCSV writes the provided records as CSV rows to the underlying writer.
// The header row contains: service, checked_at, drifted, drifted_keys.
func (e *Exporter) ExportCSV(records []Record) error {
	cw := csv.NewWriter(e.w)

	if err := cw.Write([]string{"service", "checked_at", "drifted", "drifted_keys"}); err != nil {
		return fmt.Errorf("write header: %w", err)
	}

	for _, r := range records {
		keys := buildKeyList(r)
		row := []string{
			r.ServiceName,
			r.CheckedAt.UTC().Format(time.RFC3339),
			strconv.FormatBool(r.Result.HasDrift()),
			keys,
		}
		if err := cw.Write(row); err != nil {
			return fmt.Errorf("write row for %s: %w", r.ServiceName, err)
		}
	}

	cw.Flush()
	return cw.Error()
}

// buildKeyList returns a semicolon-separated list of drifted key names.
func buildKeyList(r Record) string {
	if !r.Result.HasDrift() {
		return ""
	}
	out := ""
	for i, d := range r.Result.Diffs {
		if i > 0 {
			out += ";"
		}
		out += d.Key
	}
	return out
}
