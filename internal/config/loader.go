package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// ServiceConfig represents the declared state of a service.
type ServiceConfig struct {
	ServiceName string            `json:"service_name"`
	Version     string            `json:"version"`
	Properties  map[string]string `json:"properties"`
}

// Loader reads and parses service configuration files.
type Loader struct {
	basePath string
}

// NewLoader creates a new Loader rooted at basePath.
func NewLoader(basePath string) *Loader {
	return &Loader{basePath: basePath}
}

// Load reads a JSON config file for the given service name.
func (l *Loader) Load(serviceName string) (*ServiceConfig, error) {
	filePath := fmt.Sprintf("%s/%s.json", l.basePath, serviceName)

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("config: failed to read file %q: %w", filePath, err)
	}

	var cfg ServiceConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: failed to parse file %q: %w", filePath, err)
	}

	if cfg.ServiceName == "" {
		return nil, fmt.Errorf("config: service_name is required in %q", filePath)
	}

	return &cfg, nil
}

// LoadAll reads all JSON config files in basePath.
func (l *Loader) LoadAll() ([]*ServiceConfig, error) {
	entries, err := os.ReadDir(l.basePath)
	if err != nil {
		return nil, fmt.Errorf("config: failed to read directory %q: %w", l.basePath, err)
	}

	var configs []*ServiceConfig
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if len(name) < 5 || name[len(name)-5:] != ".json" {
			continue
		}
		serviceName := name[:len(name)-5]
		cfg, err := l.Load(serviceName)
		if err != nil {
			return nil, err
		}
		configs = append(configs, cfg)
	}
	return configs, nil
}
