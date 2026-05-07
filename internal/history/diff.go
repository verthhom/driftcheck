// Package history manages drift history records, querying, and reporting.
package history

import (
	"sort"
	"time"
)

// DiffEntry represents a single key that changed between two drift records.
type DiffEntry struct {
	Key      string
	Prev     string
	Curr     string
	Appeared bool // true if the key was absent in the previous record
	Vanished bool // true if the key is absent in the current record
}

// DiffResult holds the comparison between two consecutive drift records.
type DiffResult struct {
	ServiceName string
	PrevTime    time.Time
	CurrTime    time.Time
	Changes     []DiffEntry
}

// HasChanges returns true when at least one key difference was detected.
func (d DiffResult) HasChanges() bool {
	return len(d.Changes) > 0
}

// Diff compares two consecutive DriftRecords for the same service and returns
// the keys whose expected/actual values changed between the two snapshots.
// prev and curr must belong to the same service; no error is returned for
// mismatched service names — callers are responsible for pairing records.
func Diff(prev, curr DriftRecord) DiffResult {
	result := DiffResult{
		ServiceName: curr.ServiceName,
		PrevTime:    prev.CapturedAt,
		CurrTime:    curr.CapturedAt,
	}

	prevDrifted := indexDrifted(prev)
	currDrifted := indexDrifted(curr)

	// Keys present in prev
	for key, prevEntry := range prevDrifted {
		currEntry, exists := currDrifted[key]
		if !exists {
			result.Changes = append(result.Changes, DiffEntry{
				Key:      key,
				Prev:     prevEntry,
				Vanished: true,
			})
			continue
		}
		if currEntry != prevEntry {
			result.Changes = append(result.Changes, DiffEntry{
				Key:  key,
				Prev: prevEntry,
				Curr: currEntry,
			})
		}
	}

	// Keys new in curr
	for key, currEntry := range currDrifted {
		if _, exists := prevDrifted[key]; !exists {
			result.Changes = append(result.Changes, DiffEntry{
				Key:      key,
				Curr:     currEntry,
				Appeared: true,
			})
		}
	}

	sort.Slice(result.Changes, func(i, j int) bool {
		return result.Changes[i].Key < result.Changes[j].Key
	})

	return result
}

// indexDrifted builds a map of key -> actual value for all drifted keys in a record.
func indexDrifted(r DriftRecord) map[string]string {
	m := make(map[string]string, len(r.Drifted))
	for _, d := range r.Drifted {
		m[d.Key] = d.Actual
	}
	return m
}
