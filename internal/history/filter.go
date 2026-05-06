// Package history manages drift check records, querying, and reporting.
package history

import "time"

// FilterOptions controls which records are returned by Filter.
type FilterOptions struct {
	// Service restricts results to a specific service name.
	// An empty string matches all services.
	Service string

	// Since discards records captured before this time.
	// A zero value imposes no lower bound.
	Since time.Time

	// Until discards records captured after this time.
	// A zero value imposes no upper bound.
	Until time.Time

	// OnlyDrifted, when true, returns only records that contain at least one
	// drifted key.
	OnlyDrifted bool
}

// Filter returns the subset of records that satisfy all conditions in opts.
func Filter(records []Record, opts FilterOptions) []Record {
	out := make([]Record, 0, len(records))
	for _, r := range records {
		if opts.Service != "" && r.ServiceName != opts.Service {
			continue
		}
		if !opts.Since.IsZero() && r.CapturedAt.Before(opts.Since) {
			continue
		}
		if !opts.Until.IsZero() && r.CapturedAt.After(opts.Until) {
			continue
		}
		if opts.OnlyDrifted && len(r.Diffs) == 0 {
			continue
		}
		out = append(out, r)
	}
	return out
}
