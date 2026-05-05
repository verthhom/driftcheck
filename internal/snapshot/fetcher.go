package snapshot

import (
	"fmt"
	"os"
	"strings"
)

// Fetcher defines how a live service snapshot is obtained.
type Fetcher interface {
	Fetch(serviceName string) (*ServiceSnapshot, error)
}

// EnvFetcher reads configuration from environment variables whose names
// are prefixed with "<SERVICE_NAME>_" (upper-cased).
// This is useful for local development and testing.
type EnvFetcher struct{}

// NewEnvFetcher returns a Fetcher backed by environment variables.
func NewEnvFetcher() *EnvFetcher {
	return &EnvFetcher{}
}

// Fetch scans the environment for variables prefixed with
// "<SERVICENAME>_" and returns them as a ServiceSnapshot.
func (f *EnvFetcher) Fetch(serviceName string) (*ServiceSnapshot, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("snapshot: service name must not be empty")
	}

	prefix := strings.ToUpper(serviceName) + "_"
	config := make(map[string]string)

	for _, entry := range os.Environ() {
		parts := strings.SplitN(entry, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key, value := parts[0], parts[1]
		if strings.HasPrefix(key, prefix) {
			// Strip the service prefix to get the bare config key.
			bare := strings.TrimPrefix(key, prefix)
			config[strings.ToLower(bare)] = value
		}
	}

	return New(serviceName, config), nil
}
