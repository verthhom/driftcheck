package history

import (
	"sort"
	"time"
)

// QueryOptions controls filtering and pagination of history records.
type QueryOptions struct {
	// ServiceName filters results to a specific service. Empty means all services.
	ServiceName string
	// Since filters results to records captured after this time.
	Since time.Time
	// Limit caps the number of results returned. Zero means no limit.
	Limit int
	// OnlyDrifted filters results to only records that had drift.
	OnlyDrifted bool
}

// Query returns history records matching the given options, sorted by
// CapturedAt descending (most recent first).
func (s *Store) Query(opts QueryOptions) ([]Record, error) {
	all, err := s.List(opts.ServiceName)
	if err != nil {
		return nil, err
	}

	var filtered []Record
	for _, r := range all {
		if !opts.Since.IsZero() && !r.CapturedAt.After(opts.Since) {
			continue
		}
		if opts.OnlyDrifted && len(r.Drifts) == 0 {
			continue
		}
		filtered = append(filtered, r)
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].CapturedAt.After(filtered[j].CapturedAt)
	})

	if opts.Limit > 0 && len(filtered) > opts.Limit {
		filtered = filtered[:opts.Limit]
	}

	return filtered, nil
}

// Latest returns the most recent record for the given service name.
// Returns nil, nil if no records exist.
func (s *Store) Latest(serviceName string) (*Record, error) {
	records, err := s.Query(QueryOptions{
		ServiceName: serviceName,
		Limit:       1,
	})
	if err != nil {
		return nil, err
	}
	if len(records) == 0 {
		return nil, nil
	}
	return &records[0], nil
}

// CountDrifted returns the number of records that contain drift for the given
// service name. If serviceName is empty, all services are counted.
func (s *Store) CountDrifted(serviceName string) (int, error) {
	records, err := s.Query(QueryOptions{
		ServiceName: serviceName,
		OnlyDrifted: true,
	})
	if err != nil {
		return 0, err
	}
	return len(records), nil
}
