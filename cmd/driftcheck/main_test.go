package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// buildBinary compiles the driftcheck binary into a temp dir and returns its path.
func buildBinary(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	binPath := filepath.Join(tmpDir, "driftcheck")
	cmd := exec.Command("go", "build", "-o", binPath, ".")
	cmd.Dir = "."
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("failed to build binary: %v\n%s", err, out)
	}
	return binPath
}

func writeConfig(t *testing.T, dir, name, content string) {
	t.Helper()
	err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to write config file: %v", err)
	}
}

func TestMain_MissingConfigDir(t *testing.T) {
	bin := buildBinary(t)
	cmd := exec.Command(bin, "-config", "/nonexistent/path")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit for missing config dir")
	}
}

func TestMain_NoDrift(t *testing.T) {
	bin := buildBinary(t)
	cfgDir := t.TempDir()
	writeConfig(t, cfgDir, "svc.yaml", `
service_name: TESTSVC
config:
  TESTSVC_FOO: bar
`)
	t.Setenv("TESTSVC_FOO", "bar")

	cmd := exec.Command(bin, "-config", cfgDir)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v\noutput: %s", err, out)
	}
}

func TestMain_WithDrift(t *testing.T) {
	bin := buildBinary(t)
	cfgDir := t.TempDir()
	writeConfig(t, cfgDir, "svc.yaml", `
service_name: TESTSVC
config:
  TESTSVC_FOO: expected_value
`)
	t.Setenv("TESTSVC_FOO", "actual_value")

	cmd := exec.Command(bin, "-config", cfgDir)
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit when drift is detected")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok {
		t.Fatalf("expected ExitError, got %T", err)
	}
	if exitErr.ExitCode() != 2 {
		t.Errorf("expected exit code 2, got %d", exitErr.ExitCode())
	}
}
