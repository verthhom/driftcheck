package history_test

import (
	"os"
	"testing"
	"time"

	"driftcheck/internal/history"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "history-test-*")
	if err != nil {
		t.Fatalf("create temp dir: %v", err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })
	return dir
}

func TestSave_AndList_RoundTrip(t *testing.T) {
	store, err := history.NewStore(tempDir(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	rec := history.Record{
		ServiceName: "payments",
		CheckedAt:   time.Now().UTC().Truncate(time.Second),
		HasDrift:    true,
		Drifts:      map[string]string{"LOG_LEVEL": "want debug got info"},
	}
	if err := store.Save(rec); err != nil {
		t.Fatalf("Save: %v", err)
	}
	records, err := store.List("payments")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	got := records[0]
	if got.ServiceName != rec.ServiceName {
		t.Errorf("ServiceName: got %q, want %q", got.ServiceName, rec.ServiceName)
	}
	if !got.CheckedAt.Equal(rec.CheckedAt) {
		t.Errorf("CheckedAt: got %v, want %v", got.CheckedAt, rec.CheckedAt)
	}
	if got.HasDrift != rec.HasDrift {
		t.Errorf("HasDrift: got %v, want %v", got.HasDrift, rec.HasDrift)
	}
}

func TestSave_EmptyServiceName(t *testing.T) {
	store, _ := history.NewStore(tempDir(t))
	err := store.Save(history.Record{CheckedAt: time.Now()})
	if err == nil {
		t.Fatal("expected error for empty service name, got nil")
	}
}

func TestList_NoRecords(t *testing.T) {
	store, _ := history.NewStore(tempDir(t))
	records, err := store.List("unknown-service")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestList_MultipleRecords(t *testing.T) {
	store, _ := history.NewStore(tempDir(t))
	for i := 0; i < 3; i++ {
		err := store.Save(history.Record{
			ServiceName: "auth",
			CheckedAt:   time.Now().Add(time.Duration(i) * time.Millisecond),
			HasDrift:    i%2 == 0,
		})
		if err != nil {
			t.Fatalf("Save %d: %v", i, err)
		}
	}
	records, err := store.List("auth")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}
