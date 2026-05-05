// Package history provides a simple file-backed store for persisting
// drift check results over time.
//
// Each check result is saved as an individual JSON file named after the
// service and the Unix nanosecond timestamp of the check. Records can
// be listed per service to observe how drift evolves across deployments.
//
// Typical usage:
//
//	store, err := history.NewStore("/var/lib/driftcheck/history")
//	if err != nil { ... }
//
//	err = store.Save(history.Record{
//		ServiceName: "payments",
//		CheckedAt:   time.Now(),
//		HasDrift:    true,
//		Drifts:      map[string]string{"LOG_LEVEL": "want debug got info"},
//	})
//
//	records, err := store.List("payments")
package history
