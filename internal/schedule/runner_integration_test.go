package schedule_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"driftcheck/internal/config"
	"driftcheck/internal/history"
	"driftcheck/internal/schedule"
)

// TestRunner_PersistsHistoryRecord verifies that a completed run writes a
// history record to the store directory.
func TestRunner_PersistsHistoryRecord(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir configs: %v", err)
	}

	// Write a minimal service config.
	cfgPath := filepath.Join(cfgDir, "svc.yaml")
	writeRunnerConfig(t, cfgPath, "svc", map[string]string{"APP_ENV": "production"})

	// Set the environment variable so the snapshot matches.
	t.Setenv("APP_ENV", "production")

	histDir := filepath.Join(dir, "history")
	store, err := history.NewStore(histDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	loader := config.NewLoader()
	var buf bytes.Buffer

	runner, err := schedule.NewRunner(schedule.RunnerConfig{
		ConfigDir: cfgDir,
		Loader:    loader,
		Store:     store,
		Output:    &buf,
	})
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := runner.RunOnce(ctx); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	records, err := store.List("svc")
	if err != nil {
		t.Fatalf("store.List: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected at least one history record after RunOnce, got none")
	}
	if records[0].ServiceName != "svc" {
		t.Errorf("ServiceName = %q; want %q", records[0].ServiceName, "svc")
	}
	if records[0].HasDrift {
		t.Error("expected no drift for matching config")
	}
}

// TestRunner_RecordsCapturedDrift ensures that a drift mismatch is reflected
// correctly in the persisted history record.
func TestRunner_RecordsCapturedDrift(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	cfgDir := filepath.Join(dir, "configs")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatalf("mkdir configs: %v", err)
	}

	// Declared value differs from what we set in the environment.
	cfgPath := filepath.Join(cfgDir, "svc2.yaml")
	writeRunnerConfig(t, cfgPath, "svc2", map[string]string{"APP_ENV": "staging"})
	t.Setenv("APP_ENV", "production")

	histDir := filepath.Join(dir, "history")
	store, err := history.NewStore(histDir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	loader := config.NewLoader()
	var buf bytes.Buffer

	runner, err := schedule.NewRunner(schedule.RunnerConfig{
		ConfigDir: cfgDir,
		Loader:    loader,
		Store:     store,
		Output:    &buf,
	})
	if err != nil {
		t.Fatalf("NewRunner: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := runner.RunOnce(ctx); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	records, err := store.List("svc2")
	if err != nil {
		t.Fatalf("store.List: %v", err)
	}
	if len(records) == 0 {
		t.Fatal("expected at least one history record, got none")
	}
	if !records[0].HasDrift {
		t.Error("expected HasDrift=true for mismatched config")
	}
}

// writeRunnerConfig writes a YAML service config file for integration tests.
func writeRunnerConfig(t *testing.T, path, serviceName string, vars map[string]string) {
	t.Helper()
	content := fmt.Sprintf("service_name: %s\nenv:\n", serviceName)
	for k, v := range vars {
		content += fmt.Sprintf("  %s: %q\n", k, v)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write config %s: %v", path, err)
	}
}
