// Package drift provides the core logic for detecting and reporting
// configuration drift between a declared service configuration and a
// live snapshot of that service's running environment.
//
// Typical usage:
//
//	// Build a detector from a loaded config and a live snapshot.
//	detector := drift.NewDetector(cfg, snap)
//	result, err := detector.Check()
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Report the result to stdout.
//	reporter := drift.NewReporter(nil)
//	reporter.Report(result)
//
// The Result type carries all drifted keys along with their expected
// and actual values and a human-readable reason string, making it
// straightforward to render reports in different formats.
package drift
