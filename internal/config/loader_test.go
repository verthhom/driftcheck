package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"driftcheck/internal/config"
)

func writeTempConfig(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}
}

func TestLoad_ValidConfig(t *testing.T) {
	dir := t.TempDir()
	writeTempConfig(t, dir, "api.json", `{
		"service_name": "api",
		"version": "1.2.3",
		"properties": {"replicas": "3", "region": "us-east-1"}
	}`)

	loader := config.NewLoader(dir)
	cfg, err := loader.Load("api")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.ServiceName != "api" {
		t.Errorf("expected service_name 'api', got %q", cfg.ServiceName)
	}
	if cfg.Version != "1.2.3" {
		t.Errorf("expected version '1.2.3', got %q", cfg.Version)
	}
	if cfg.Properties["replicas"] != "3" {
		t.Errorf("expected replicas '3', got %q", cfg.Properties["replicas"])
	}
}

func TestLoad_MissingFile(t *testing.T) {
	loader := config.NewLoader(t.TempDir())
	_, err := loader.Load("nonexistent")
	if err == nil {
		t.Fatal("expected error for missing file, got nil")
	}
}

func TestLoad_MissingServiceName(t *testing.T) {
	dir := t.TempDir()
	writeTempConfig(t, dir, "bad.json", `{"version": "1.0.0", "properties": {}}`)

	loader := config.NewLoader(dir)
	_, err := loader.Load("bad")
	if err == nil {
		t.Fatal("expected error for missing service_name, got nil")
	}
}

func TestLoadAll_MultipleConfigs(t *testing.T) {
	dir := t.TempDir()
	writeTempConfig(t, dir, "svc-a.json", `{"service_name": "svc-a", "version": "1.0", "properties": {}}`)
	writeTempConfig(t, dir, "svc-b.json", `{"service_name": "svc-b", "version": "2.0", "properties": {}}`)
	writeTempConfig(t, dir, "notes.txt", "not a config")

	loader := config.NewLoader(dir)
	cfgs, err := loader.LoadAll()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(cfgs) != 2 {
		t.Errorf("expected 2 configs, got %d", len(cfgs))
	}
}
