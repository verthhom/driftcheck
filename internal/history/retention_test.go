package history

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeFile(t *testing.T, dir, name string, modTime time.Time) {
	t.Helper()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(`{}`), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
}

func TestPurge_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	n, err := Purge(dir, RetentionPolicy{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deletions, got %d", n)
	}
}

func TestPurge_NonExistentDir(t *testing.T) {
	n, err := Purge("/tmp/driftcheck_no_such_dir_xyz", RetentionPolicy{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Errorf("expected 0 deletions, got %d", n)
	}
}

func TestPurge_MaxAge(t *testing.T) {
	dir := t.TempDir()
	old := time.Now().Add(-48 * time.Hour)
	recent := time.Now().Add(-1 * time.Minute)

	writeFile(t, dir, "svc_2024-01-01T00:00:00Z.json", old)
	writeFile(t, dir, "svc_2024-06-01T00:00:00Z.json", recent)

	n, err := Purge(dir, RetentionPolicy{MaxAge: 24 * time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 deletion, got %d", n)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Errorf("expected 1 remaining file, got %d", len(entries))
	}
}

func TestPurge_MaxRecordsPerService(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	for _, name := range []string{
		"api_2024-01-01T00:00:00Z.json",
		"api_2024-02-01T00:00:00Z.json",
		"api_2024-03-01T00:00:00Z.json",
	} {
		writeFile(t, dir, name, now)
	}

	n, err := Purge(dir, RetentionPolicy{MaxRecordsPerService: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 1 {
		t.Errorf("expected 1 deletion, got %d", n)
	}
}

func TestPurge_MultipleServices(t *testing.T) {
	dir := t.TempDir()
	now := time.Now()

	for _, name := range []string{
		"alpha_2024-01-01T00:00:00Z.json",
		"alpha_2024-02-01T00:00:00Z.json",
		"beta_2024-01-01T00:00:00Z.json",
		"beta_2024-02-01T00:00:00Z.json",
	} {
		writeFile(t, dir, name, now)
	}

	n, err := Purge(dir, RetentionPolicy{MaxRecordsPerService: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 2 {
		t.Errorf("expected 2 deletions, got %d", n)
	}
}

func TestServiceFromFilename(t *testing.T) {
	cases := []struct {
		input string
		want  string
	}{
		{"my-service_2024-01-01T00:00:00Z.json", "my-service"},
		{"svc_ts.json", "svc"},
		{"nounderscore.json", "nounderscore"},
	}
	for _, tc := range cases {
		got := serviceFromFilename(tc.input)
		if got != tc.want {
			t.Errorf("serviceFromFilename(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}
