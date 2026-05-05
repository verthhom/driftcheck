// Package config provides types and utilities for loading declared service
// configurations from JSON files on disk.
//
// A ServiceConfig describes the intended (desired) state of a service,
// including its name, version, and arbitrary key-value properties.
// The Loader is responsible for reading these configs so they can be
// compared against live state by the drift detector.
//
// Typical usage:
//
//	loader := config.NewLoader("/etc/driftcheck/configs")
//	cfg, err := loader.Load("my-service")
//	if err != nil {
//		log.Fatal(err)
//	}
//	// pass cfg to drift.Detector for comparison
package config
