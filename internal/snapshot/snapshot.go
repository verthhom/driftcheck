// Package snapshot provides types and utilities for capturing and
// representing the live state of a deployed service.
package snapshot

import "time"

// ServiceSnapshot holds the observed runtime configuration of a service
// at a specific point in time.
type ServiceSnapshot struct {
	// ServiceName is the canonical name of the deployed service.
	ServiceName string `json:"service_name"`

	// Config holds the key/value pairs observed from the live service.
	Config map[string]string `json:"config"`

	// CapturedAt is the UTC timestamp when the snapshot was taken.
	CapturedAt time.Time `json:"captured_at"`
}

// New creates a new ServiceSnapshot with CapturedAt set to the current UTC time.
func New(serviceName string, config map[string]string) *ServiceSnapshot {
	copy := make(map[string]string, len(config))
	for k, v := range config {
		copy[k] = v
	}
	return &ServiceSnapshot{
		ServiceName: serviceName,
		Config:      copy,
		CapturedAt:  time.Now().UTC(),
	}
}

// Get returns the value for the given key and whether it was present.
func (s *ServiceSnapshot) Get(key string) (string, bool) {
	v, ok := s.Config[key]
	return v, ok
}
