// Package history provides persistence, querying, and analysis of drift check
// records produced by the driftcheck runner.
//
// Core types:
//
//   - DriftRecord  — a point-in-time record of a single service check.
//   - DriftedKey   — a key whose live value differs from its declared value.
//
// Sub-features:
//
//   - Store      — save and list records on disk as JSON files.
//   - Query      — filter and retrieve records from a store.
//   - Filter     — in-memory slice filtering by service, time range, or drift.
//   - Summarise  — aggregate statistics across a slice of records.
//   - Diff       — compare two consecutive records to surface key-level changes.
//   - Exporter   — write records to CSV for external analysis.
//   - Retention  — purge old records by age or per-service count.
//   - Cleanup    — remove history files for services no longer being monitored.
package history
