package drift_test

import (
	"testing"

	"github.com/yourorg/driftcheck/internal/drift"
)

func TestCheck_NoDrift(t *testing.T) {
	detector := drift.NewDetector()
	declared := drift.State{
		ServiceName: "api",
		Config:      map[string]string{"PORT": "8080", "LOG_LEVEL": "info"},
	}
	deployed := drift.State{
		ServiceName: "api",
		Config:      map[string]string{"PORT": "8080", "LOG_LEVEL": "info"},
	}

	result, err := detector.Check(declared, deployed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Drifted {
		t.Errorf("expected no drift, got diffs: %+v", result.Diffs)
	}
}

func TestCheck_ValueMismatch(t *testing.T) {
	detector := drift.NewDetector()
	declared := drift.State{
		ServiceName: "api",
		Config:      map[string]string{"PORT": "8080"},
	}
	deployed := drift.State{
		ServiceName: "api",
		Config:      map[string]string{"PORT": "9090"},
	}

	result, err := detector.Check(declared, deployed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Drifted {
		t.Fatal("expected drift but got none")
	}
	if len(result.Diffs) != 1 || result.Diffs[0].Key != "PORT" {
		t.Errorf("unexpected diffs: %+v", result.Diffs)
	}
}

func TestCheck_MissingKey(t *testing.T) {
	detector := drift.NewDetector()
	declared := drift.State{
		ServiceName: "api",
		Config:      map[string]string{"PORT": "8080", "TIMEOUT": "30s"},
	}
	deployed := drift.State{
		ServiceName: "api",
		Config:      map[string]string{"PORT": "8080"},
	}

	result, err := detector.Check(declared, deployed)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Drifted {
		t.Fatal("expected drift due to missing key")
	}
}

func TestCheck_ServiceNameMismatch(t *testing.T) {
	detector := drift.NewDetector()
	declared := drift.State{ServiceName: "api", Config: map[string]string{}}
	deployed := drift.State{ServiceName: "worker", Config: map[string]string{}}

	_, err := detector.Check(declared, deployed)
	if err == nil {
		t.Fatal("expected error for service name mismatch")
	}
}
