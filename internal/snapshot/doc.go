// Package snapshot captures and represents the live (deployed) state of a
// service so that it can be compared against the declared configuration.
//
// # Overview
//
// A [ServiceSnapshot] is an immutable point-in-time view of a service's
// runtime configuration. Snapshots are produced by a [Fetcher].
//
// # Fetchers
//
// The package ships with [EnvFetcher], which derives configuration from
// environment variables. Custom fetchers (e.g. Kubernetes ConfigMap readers,
// AWS Parameter Store clients) can be added by implementing the [Fetcher]
// interface.
package snapshot
