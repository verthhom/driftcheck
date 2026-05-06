package history

import (
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestQuery_FilterByService(t *testing.T) {
	dir := tempDir(t)
	s := NewStore(dir)

	save := func(svc string, drifts []string) {
		err := s.Save(Record{ServiceName: svc, CapturedAt: time.Now(), Drifts: drifts})
		if err != nil {
			t.Fatalf("Save: %v", err)
		}
	}

	save("alpha", nil)
	save("beta", []string{"KEY_A"})
	save("alpha", []string{"KEY_B"})

	results, err := s.Query(QueryOptions{ServiceName: "alpha"})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 records for alpha, got %d", len(results))
	}
	for _, r := range results {
		if r.ServiceName != "alpha" {
			t.Errorf("unexpected service name %q", r.ServiceName)
		}
	}
}

func TestQuery_OnlyDrifted(t *testing.T) {
	dir := tempDir(t)
	s := NewStore(dir)

	_ = s.Save(Record{ServiceName: "svc", CapturedAt: time.Now(), Drifts: nil})
	_ = s.Save(Record{ServiceName: "svc", CapturedAt: time.Now(), Drifts: []string{"X"}})

	results, err := s.Query(QueryOptions{ServiceName: "svc", OnlyDrifted: true})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 drifted record, got %d", len(results))
	}
	if diff := cmp.Diff([]string{"X"}, results[0].Drifts); diff != "" {
		t.Errorf("drifts mismatch (-want +got):\n%s", diff)
	}
}

func TestQuery_Limit(t *testing.T) {
	dir := tempDir(t)
	s := NewStore(dir)

	for i := 0; i < 5; i++ {
		_ = s.Save(Record{ServiceName: "svc", CapturedAt: time.Now()})
	}

	results, err := s.Query(QueryOptions{ServiceName: "svc", Limit: 3})
	if err != nil {
		t.Fatalf("Query: %v", err)
	}
	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
}

func TestLatest_ReturnsNilWhenEmpty(t *testing.T) {
	dir := tempDir(t)
	s := NewStore(dir)

	r, err := s.Latest("nonexistent")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if r != nil {
		t.Errorf("expected nil, got %+v", r)
	}
}

func TestLatest_ReturnsMostRecent(t *testing.T) {
	dir := tempDir(t)
	s := NewStore(dir)

	old := time.Now().Add(-time.Hour)
	recent := time.Now()

	_ = s.Save(Record{ServiceName: "svc", CapturedAt: old, Drifts: nil})
	_ = s.Save(Record{ServiceName: "svc", CapturedAt: recent, Drifts: []string{"LATEST"}})

	r, err := s.Latest("svc")
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if r == nil {
		t.Fatal("expected a record, got nil")
	}
	if diff := cmp.Diff([]string{"LATEST"}, r.Drifts); diff != "" {
		t.Errorf("drifts mismatch (-want +got):\n%s", diff)
	}
}
