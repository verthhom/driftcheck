package history

import (
	"testing"
	"time"
)

func makeRecord(service string, hasDrift bool, at time.Time) DriftRecord {
	return DriftRecord{
		ServiceName: service,
		HasDrift:    hasDrift,
		CapturedAt:  at,
	}
}

func TestSummarise_EmptyInput(t *testing.T) {
	result := Summarise(nil)
	if len(result) != 0 {
		t.Fatalf("expected empty summaries, got %d", len(result))
	}
}

func TestSummarise_SingleServiceNoDrift(t *testing.T) {
	now := time.Now()
	records := []DriftRecord{
		makeRecord("svc-a", false, now.Add(-2*time.Minute)),
		makeRecord("svc-a", false, now.Add(-1*time.Minute)),
	}

	summaries := Summarise(records)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.ServiceName != "svc-a" {
		t.Errorf("unexpected service name: %s", s.ServiceName)
	}
	if s.TotalChecks != 2 {
		t.Errorf("expected TotalChecks=2, got %d", s.TotalChecks)
	}
	if s.DriftedChecks != 0 {
		t.Errorf("expected DriftedChecks=0, got %d", s.DriftedChecks)
	}
	if s.DriftRate != 0.0 {
		t.Errorf("expected DriftRate=0, got %f", s.DriftRate)
	}
	if s.LastDriftedAt != nil {
		t.Error("expected LastDriftedAt to be nil")
	}
}

func TestSummarise_SingleServiceWithDrift(t *testing.T) {
	now := time.Now()
	records := []DriftRecord{
		makeRecord("svc-b", false, now.Add(-3*time.Minute)),
		makeRecord("svc-b", true, now.Add(-2*time.Minute)),
		makeRecord("svc-b", true, now.Add(-1*time.Minute)),
	}

	summaries := Summarise(records)
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	s := summaries[0]
	if s.DriftedChecks != 2 {
		t.Errorf("expected DriftedChecks=2, got %d", s.DriftedChecks)
	}
	expectedRate := 2.0 / 3.0 * 100.0
	if s.DriftRate < expectedRate-0.001 || s.DriftRate > expectedRate+0.001 {
		t.Errorf("expected DriftRate~=%.4f, got %.4f", expectedRate, s.DriftRate)
	}
	if s.LastDriftedAt == nil {
		t.Fatal("expected LastDriftedAt to be set")
	}
	if !s.LastDriftedAt.Equal(now.Add(-1 * time.Minute)) {
		t.Errorf("LastDriftedAt not pointing to most recent drift")
	}
}

func TestSummarise_MultipleServices(t *testing.T) {
	now := time.Now()
	records := []DriftRecord{
		makeRecord("alpha", true, now.Add(-5*time.Minute)),
		makeRecord("beta", false, now.Add(-4*time.Minute)),
		makeRecord("alpha", false, now.Add(-3*time.Minute)),
	}

	summaries := Summarise(records)
	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}

	byName := make(map[string]ServiceSummary)
	for _, s := range summaries {
		byName[s.ServiceName] = s
	}

	if byName["alpha"].TotalChecks != 2 {
		t.Errorf("alpha: expected TotalChecks=2, got %d", byName["alpha"].TotalChecks)
	}
	if byName["beta"].DriftedChecks != 0 {
		t.Errorf("beta: expected DriftedChecks=0, got %d", byName["beta"].DriftedChecks)
	}
}
