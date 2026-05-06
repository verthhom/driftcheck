package history

import (
	"bytes"
	"encoding/csv"
	"strings"
	"testing"
	"time"

	"driftcheck/internal/drift"
)

func makeExportRecord(service string, diffs []drift.Diff) Record {
	return Record{
		ServiceName: service,
		CheckedAt:   time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC),
		Result:      drift.Result{Diffs: diffs},
	}
}

func TestExportCSV_Header(t *testing.T) {
	var buf bytes.Buffer
	e := NewExporter(&buf)
	if err := e.ExportCSV(nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := csv.NewReader(&buf)
	rows, err := r.ReadAll()
	if err != nil {
		t.Fatalf("csv parse error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row (header), got %d", len(rows))
	}
	expected := []string{"service", "checked_at", "drifted", "drifted_keys"}
	for i, col := range expected {
		if rows[0][i] != col {
			t.Errorf("col %d: want %q got %q", i, col, rows[0][i])
		}
	}
}

func TestExportCSV_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	e := NewExporter(&buf)
	recs := []Record{makeExportRecord("svc-a", nil)}
	if err := e.ExportCSV(recs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := csv.NewReader(&buf)
	rows, _ := r.ReadAll()
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[1][0] != "svc-a" {
		t.Errorf("service: want svc-a got %s", rows[1][0])
	}
	if rows[1][2] != "false" {
		t.Errorf("drifted: want false got %s", rows[1][2])
	}
	if rows[1][3] != "" {
		t.Errorf("drifted_keys: want empty got %s", rows[1][3])
	}
}

func TestExportCSV_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	e := NewExporter(&buf)
	diffs := []drift.Diff{
		{Key: "PORT", Expected: "8080", Actual: "9090"},
		{Key: "LOG_LEVEL", Expected: "info", Actual: "debug"},
	}
	recs := []Record{makeExportRecord("svc-b", diffs)}
	if err := e.ExportCSV(recs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := csv.NewReader(&buf)
	rows, _ := r.ReadAll()
	if rows[1][2] != "true" {
		t.Errorf("drifted: want true got %s", rows[1][2])
	}
	if !strings.Contains(rows[1][3], "PORT") || !strings.Contains(rows[1][3], "LOG_LEVEL") {
		t.Errorf("drifted_keys missing expected keys: %s", rows[1][3])
	}
}

func TestExportCSV_MultipleRecords(t *testing.T) {
	var buf bytes.Buffer
	e := NewExporter(&buf)
	recs := []Record{
		makeExportRecord("svc-a", nil),
		makeExportRecord("svc-b", []drift.Diff{{Key: "X", Expected: "1", Actual: "2"}}),
	}
	if err := e.ExportCSV(recs); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	r := csv.NewReader(&buf)
	rows, _ := r.ReadAll()
	if len(rows) != 3 {
		t.Fatalf("expected 3 rows, got %d", len(rows))
	}
}
