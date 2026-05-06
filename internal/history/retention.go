package history

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// RetentionPolicy defines how old records are pruned from the history store.
type RetentionPolicy struct {
	// MaxAge is the maximum age of records to keep. Zero means no age limit.
	MaxAge time.Duration
	// MaxRecordsPerService is the maximum number of records to keep per service.
	// Zero means no limit.
	MaxRecordsPerService int
}

// Purge removes records from dir that violate the given policy.
// It returns the number of files deleted and any error encountered.
func Purge(dir string, policy RetentionPolicy) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, nil
		}
		return 0, fmt.Errorf("read history dir: %w", err)
	}

	// Group JSON files by service name.
	byService := make(map[string][]os.DirEntry)
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		service := serviceFromFilename(e.Name())
		byService[service] = append(byService[service], e)
	}

	now := time.Now()
	deleted := 0

	for _, files := range byService {
		// Sort oldest first by filename (filenames embed RFC3339 timestamps).
		sort.Slice(files, func(i, j int) bool {
			return files[i].Name() < files[j].Name()
		})

		for idx, f := range files {
			remove := false

			if policy.MaxAge > 0 {
				info, err := f.Info()
				if err == nil && now.Sub(info.ModTime()) > policy.MaxAge {
					remove = true
				}
			}

			if policy.MaxRecordsPerService > 0 {
				excess := len(files) - policy.MaxRecordsPerService
				if excess > 0 && idx < excess {
					remove = true
				}
			}

			if remove {
				path := filepath.Join(dir, f.Name())
				if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
					return deleted, fmt.Errorf("remove %s: %w", path, err)
				}
				deleted++
			}
		}
	}

	return deleted, nil
}

// serviceFromFilename extracts the service name from a history filename.
// Expected format: "<service>_<timestamp>.json".
func serviceFromFilename(name string) string {
	base := strings.TrimSuffix(name, ".json")
	if idx := strings.LastIndex(base, "_"); idx >= 0 {
		return base[:idx]
	}
	return base
}
