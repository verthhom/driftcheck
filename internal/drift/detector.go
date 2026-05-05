package drift

import (
	"fmt"
)

// State represents the configuration state of a service.
type State struct {
	ServiceName string
	Config      map[string]string
}

// DriftResult holds the result of a drift check between declared and deployed states.
type DriftResult struct {
	ServiceName string
	Drifted     bool
	Diffs       []Diff
}

// Diff describes a single key-level difference between declared and deployed config.
type Diff struct {
	Key      string
	Declared string
	Deployed string
}

// Detector compares declared state against deployed state.
type Detector struct{}

// NewDetector creates a new Detector instance.
func NewDetector() *Detector {
	return &Detector{}
}

// Check compares declared and deployed states and returns a DriftResult.
func (d *Detector) Check(declared, deployed State) (DriftResult, error) {
	if declared.ServiceName != deployed.ServiceName {
		return DriftResult{}, fmt.Errorf(
			"service name mismatch: declared=%q deployed=%q",
			declared.ServiceName, deployed.ServiceName,
		)
	}

	result := DriftResult{
		ServiceName: declared.ServiceName,
	}

	seen := make(map[string]bool)

	for key, declaredVal := range declared.Config {
		seen[key] = true
		deployedVal, exists := deployed.Config[key]
		if !exists {
			result.Diffs = append(result.Diffs, Diff{Key: key, Declared: declaredVal, Deployed: "<missing>"})
		} else if declaredVal != deployedVal {
			result.Diffs = append(result.Diffs, Diff{Key: key, Declared: declaredVal, Deployed: deployedVal})
		}
	}

	for key, deployedVal := range deployed.Config {
		if !seen[key] {
			result.Diffs = append(result.Diffs, Diff{Key: key, Declared: "<missing>", Deployed: deployedVal})
		}
	}

	result.Drifted = len(result.Diffs) > 0
	return result, nil
}
