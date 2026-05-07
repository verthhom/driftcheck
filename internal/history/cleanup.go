// Package history manages drift check history records.
package history

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// CleanupOptions controls how orphaned history records are removed.
type CleanupOptions struct {
	// KnownServices is the set of service names that are still active.
	// Records belonging to any other service will be removed.
	KnownServices []string

	// Logger receives informational messages; defaults to stderr.
	Logger *log.Logger
}

// CleanupResult summarises what was removed.
type CleanupResult struct {
	Removed int
	Errors  []error
}

// Cleanup removes history records for services that are no longer in
// KnownServices. It is safe to call on an empty or non-existent directory.
func Cleanup(dir string, opts CleanupOptions) (CleanupResult, error) {
	logger := opts.Logger
	if logger == nil {
		logger = log.New(io.Discard, "", 0)
	}

	known := make(map[string]bool, len(opts.KnownServices))
	for _, s := range opts.KnownServices {
		known[s] = true
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return CleanupResult{}, nil
		}
		return CleanupResult{}, fmt.Errorf("cleanup: read dir: %w", err)
	}

	var result CleanupResult
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		svc := serviceFromFilename(e.Name())
		if known[svc] {
			continue
		}
		path := filepath.Join(dir, e.Name())
		if removeErr := os.Remove(path); removeErr != nil {
			result.Errors = append(result.Errors, removeErr)
			continue
		}
		logger.Printf("cleanup: removed orphaned record %s (service %q)", e.Name(), svc)
		result.Removed++
	}
	return result, nil
}
