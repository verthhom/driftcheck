package history

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCleanup_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	res, err := Cleanup(dir, CleanupOptions{KnownServices: []string{"svc-a"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 0 {
		t.Errorf("expected 0 removed, got %d", res.Removed)
	}
}

func TestCleanup_NonExistentDir(t *testing.T) {
	res, err := Cleanup("/tmp/driftcheck-does-not-exist-xyz", CleanupOptions{})
	if err != nil {
		t.Fatalf("expected no error for missing dir, got %v", err)
	}
	if res.Removed != 0 {
		t.Errorf("expected 0 removed, got %d", res.Removed)
	}
}

func TestCleanup_RemovesOrphanedService(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "svc-gone_2024-01-01T00:00:00Z.json"), `{}`)
	writeFile(t, filepath.Join(dir, "svc-kept_2024-01-01T00:00:00Z.json"), `{}`)

	res, err := Cleanup(dir, CleanupOptions{KnownServices: []string{"svc-kept"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 1 {
		t.Errorf("expected 1 removed, got %d", res.Removed)
	}
	if _, statErr := os.Stat(filepath.Join(dir, "svc-gone_2024-01-01T00:00:00Z.json")); !os.IsNotExist(statErr) {
		t.Error("orphaned file should have been deleted")
	}
	if _, statErr := os.Stat(filepath.Join(dir, "svc-kept_2024-01-01T00:00:00Z.json")); statErr != nil {
		t.Error("known-service file should still exist")
	}
}

func TestCleanup_AllKnown(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "svc-a_2024-01-01T00:00:00Z.json"), `{}`)
	writeFile(t, filepath.Join(dir, "svc-b_2024-01-01T00:00:00Z.json"), `{}`)

	res, err := Cleanup(dir, CleanupOptions{KnownServices: []string{"svc-a", "svc-b"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 0 {
		t.Errorf("expected 0 removed, got %d", res.Removed)
	}
}

func TestCleanup_IgnoresNonJSON(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "svc-gone_2024-01-01T00:00:00Z.txt"), `irrelevant`)

	res, err := Cleanup(dir, CleanupOptions{KnownServices: []string{}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Removed != 0 {
		t.Errorf("non-json files should be ignored, got removed=%d", res.Removed)
	}
}
