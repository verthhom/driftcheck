// Package history provides persistence for drift check results,
// allowing comparison of drift over time.
package history

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Record represents a single stored drift check result.
type Record struct {
	ServiceName string            `json:"service_name"`
	CheckedAt   time.Time         `json:"checked_at"`
	HasDrift    bool              `json:"has_drift"`
	Drifts      map[string]string `json:"drifts,omitempty"`
}

// Store persists and retrieves drift history records.
type Store struct {
	dir string
}

// NewStore creates a Store that reads and writes records under dir.
func NewStore(dir string) (*Store, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("history: create store dir: %w", err)
	}
	return &Store{dir: dir}, nil
}

// Save writes a Record to disk as a JSON file named by service and timestamp.
func (s *Store) Save(r Record) error {
	if r.ServiceName == "" {
		return fmt.Errorf("history: service name must not be empty")
	}
	filename := fmt.Sprintf("%s_%d.json", r.ServiceName, r.CheckedAt.UnixNano())
	path := filepath.Join(s.dir, filename)
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("history: create record file: %w", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(r); err != nil {
		return fmt.Errorf("history: encode record: %w", err)
	}
	return nil
}

// List returns all Records stored for the given service name.
func (s *Store) List(serviceName string) ([]Record, error) {
	pattern := filepath.Join(s.dir, serviceName+"_*.json")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("history: glob records: %w", err)
	}
	var records []Record
	for _, path := range matches {
		r, err := loadRecord(path)
		if err != nil {
			return nil, err
		}
		records = append(records, r)
	}
	return records, nil
}

func loadRecord(path string) (Record, error) {
	f, err := os.Open(path)
	if err != nil {
		return Record{}, fmt.Errorf("history: open record %q: %w", path, err)
	}
	defer f.Close()
	var r Record
	if err := json.NewDecoder(f).Decode(&r); err != nil {
		return Record{}, fmt.Errorf("history: decode record %q: %w", path, err)
	}
	return r, nil
}
