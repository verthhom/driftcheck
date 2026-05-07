package history

import (
	"testing"
	"time"
)

func makeDiffRecord(service string, at time.Time, drifted []DriftedKey) DriftRecord {
	return DriftRecord{
		ServiceName: service,
		CapturedAt:  at,
		Drifted:     drifted,
	}
}

var (
	t0 = time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	t1 = t0.Add(time.Hour)
)

func TestDiff_NoDriftInEither(t *testing.T) {
	prev := makeDiffRecord("svc", t0, nil)
	curr := makeDiffRecord("svc", t1, nil)
	result := Diff(prev, curr)
	if result.HasChanges() {
		t.Fatalf("expected no changes, got %v", result.Changes)
	}
}

func TestDiff_KeyAppeared(t *testing.T) {
	prev := makeDiffRecord("svc", t0, nil)
	curr := makeDiffRecord("svc", t1, []DriftedKey{{Key: "PORT", Actual: "9090"}})
	result := Diff(prev, curr)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	entry := result.Changes[0]
	if !entry.Appeared {
		t.Errorf("expected Appeared=true")
	}
	if entry.Key != "PORT" {
		t.Errorf("expected key PORT, got %s", entry.Key)
	}
}

func TestDiff_KeyVanished(t *testing.T) {
	prev := makeDiffRecord("svc", t0, []DriftedKey{{Key: "PORT", Actual: "8080"}})
	curr := makeDiffRecord("svc", t1, nil)
	result := Diff(prev, curr)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	entry := result.Changes[0]
	if !entry.Vanished {
		t.Errorf("expected Vanished=true")
	}
	if entry.Prev != "8080" {
		t.Errorf("expected Prev=8080, got %s", entry.Prev)
	}
}

func TestDiff_ValueChanged(t *testing.T) {
	prev := makeDiffRecord("svc", t0, []DriftedKey{{Key: "PORT", Actual: "8080"}})
	curr := makeDiffRecord("svc", t1, []DriftedKey{{Key: "PORT", Actual: "9090"}})
	result := Diff(prev, curr)
	if len(result.Changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(result.Changes))
	}
	entry := result.Changes[0]
	if entry.Prev != "8080" || entry.Curr != "9090" {
		t.Errorf("unexpected values: prev=%s curr=%s", entry.Prev, entry.Curr)
	}
	if entry.Appeared || entry.Vanished {
		t.Errorf("expected neither Appeared nor Vanished")
	}
}

func TestDiff_Metadata(t *testing.T) {
	prev := makeDiffRecord("my-svc", t0, nil)
	curr := makeDiffRecord("my-svc", t1, nil)
	result := Diff(prev, curr)
	if result.ServiceName != "my-svc" {
		t.Errorf("expected service my-svc, got %s", result.ServiceName)
	}
	if !result.PrevTime.Equal(t0) {
		t.Errorf("unexpected PrevTime")
	}
	if !result.CurrTime.Equal(t1) {
		t.Errorf("unexpected CurrTime")
	}
}

func TestDiff_ChangesAreSorted(t *testing.T) {
	prev := makeDiffRecord("svc", t0, nil)
	curr := makeDiffRecord("svc", t1, []DriftedKey{
		{Key: "Z_VAR", Actual: "1"},
		{Key: "A_VAR", Actual: "2"},
		{Key: "M_VAR", Actual: "3"},
	})
	result := Diff(prev, curr)
	keys := make([]string, len(result.Changes))
	for i, c := range result.Changes {
		keys[i] = c.Key
	}
	if keys[0] != "A_VAR" || keys[1] != "M_VAR" || keys[2] != "Z_VAR" {
		t.Errorf("changes not sorted: %v", keys)
	}
}
