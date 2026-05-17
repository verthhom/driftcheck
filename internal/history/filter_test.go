package history

import (
	"testing"
	"time"
)

func makeFilterRecord(service string, capturedAt time.Time, diffs map[string][2]string) Record {
	return Record{
		ServiceName: service,
		CapturedAt:  capturedAt,
		Diffs:       diffs,
	}
}

var (
	now   = time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	early = now.Add(-48 * time.Hour)
	late  = now.Add(48 * time.Hour)
)

func TestFilter_NoOptions(t *testing.T) {
	records := []Record{
		makeFilterRecord("svc-a", now, nil),
		makeFilterRecord("svc-b", early, map[string][2]string{"KEY": {"want", "got"}}),
	}
	got := Filter(records, FilterOptions{})
	if len(got) != 2 {
		t.Fatalf("expected 2 records, got %d", len(got))
	}
}

func TestFilter_ByService(t *testing.T) {
	records := []Record{
		makeFilterRecord("svc-a", now, nil),
		makeFilterRecord("svc-b", now, nil),
		makeFilterRecord("svc-a", early, nil),
	}
	got := Filter(records, FilterOptions{Service: "svc-a"})
	if len(got) != 2 {
		t.Fatalf("expected 2 records for svc-a, got %d", len(got))
	}
	for _, r := range got {
		if r.ServiceName != "svc-a" {
			t.Errorf("unexpected service %q", r.ServiceName)
		}
	}
}

func TestFilter_Since(t *testing.T) {
	records := []Record{
		makeFilterRecord("svc", early, nil),
		makeFilterRecord("svc", now, nil),
		makeFilterRecord("svc", late, nil),
	}
	got := Filter(records, FilterOptions{Since: now})
	if len(got) != 2 {
		t.Fatalf("expected 2 records at or after now, got %d", len(got))
	}
}

func TestFilter_Until(t *testing.T) {
	records := []Record{
		makeFilterRecord("svc", early, nil),
		makeFilterRecord("svc", now, nil),
		makeFilterRecord("svc", late, nil),
	}
	got := Filter(records, FilterOptions{Until: now})
	if len(got) != 2 {
		t.Fatalf("expected 2 records at or before now, got %d", len(got))
	}
}

func TestFilter_OnlyDrifted(t *testing.T) {
	records := []Record{
		makeFilterRecord("svc", now, nil),
		makeFilterRecord("svc", now, map[string][2]string{"K": {"a", "b"}}),
		makeFilterRecord("svc", now, nil),
	}
	got := Filter(records, FilterOptions{OnlyDrifted: true})
	if len(got) != 1 {
		t.Fatalf("expected 1 drifted record, got %d", len(got))
	}
}

func TestFilter_EmptyInput(t *testing.T) {
	got := Filter([]Record{}, FilterOptions{Service: "svc-a", OnlyDrifted: true})
	if len(got) != 0 {
		t.Fatalf("expected 0 records for empty input, got %d", len(got))
	}
}

func TestFilter_Combined(t *testing.T) {
	records := []Record{
		makeFilterRecord("svc-a", early, map[string][2]string{"K": {"a", "b"}}),
		makeFilterRecord("svc-a", now, map[string][2]string{"K": {"a", "b"}}),
		makeFilterRecord("svc-b", now, map[string][2]string{"K": {"a", "b"}}),
		makeFilterRecord("svc-a", now, nil),
	}
	got := Filter(records, FilterOptions{
		Service:     "svc-a",
		Since:       now,
		OnlyDrifted: true,
	})
	if len(got) != 1 {
		t.Fatalf("expected 1 record, got %d", len(got))
	}
	if got[0].ServiceName != "svc-a" {
		t.Errorf("unexpected service %q", got[0].ServiceName)
	}
}
