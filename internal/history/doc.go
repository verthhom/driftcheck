// Package history provides persistent storage and querying of drift check
// results over time.
//
// Records are written as JSON files under a configurable directory, keyed by
// service name and capture timestamp. The Store type exposes Save, List,
// Query, and Latest operations.
//
// # Storage layout
//
//	<dir>/<service-name>/<timestamp>.json
//
// # Querying
//
// QueryOptions supports filtering by service name, time range, drift
// presence, and result count limiting. Results are always returned in
// descending capture-time order (most recent first).
package history
