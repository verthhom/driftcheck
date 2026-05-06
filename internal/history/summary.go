package history

import "time"

// ServiceSummary holds aggregated drift statistics for a single service.
type ServiceSummary struct {
	ServiceName   string
	TotalChecks   int
	DriftedChecks int
	LastCheckedAt time.Time
	LastDriftedAt *time.Time
	DriftRate     float64 // percentage 0–100
}

// Summarise computes per-service drift statistics from a slice of records.
func Summarise(records []DriftRecord) []ServiceSummary {
	type bucket struct {
		total       int
		drifted     int
		lastChecked time.Time
		lastDrifted *time.Time
	}

	buckets := make(map[string]*bucket)

	for _, r := range records {
		b, ok := buckets[r.ServiceName]
		if !ok {
			b = &bucket{}
			buckets[r.ServiceName] = b
		}

		b.total++

		if r.CapturedAt.After(b.lastChecked) {
			b.lastChecked = r.CapturedAt
		}

		if r.HasDrift {
			b.drifted++
			if b.lastDrifted == nil || r.CapturedAt.After(*b.lastDrifted) {
				t := r.CapturedAt
				b.lastDrifted = &t
			}
		}
	}

	summaries := make([]ServiceSummary, 0, len(buckets))
	for name, b := range buckets {
		rate := 0.0
		if b.total > 0 {
			rate = float64(b.drifted) / float64(b.total) * 100.0
		}
		summaries = append(summaries, ServiceSummary{
			ServiceName:   name,
			TotalChecks:   b.total,
			DriftedChecks: b.drifted,
			LastCheckedAt: b.lastChecked,
			LastDriftedAt: b.lastDrifted,
			DriftRate:     rate,
		})
	}

	return summaries
}
