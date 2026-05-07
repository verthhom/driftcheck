package schedule_test

import (
	"bytes"
	"context"
	"errors"
	"log"
	"strings"
	"testing"
	"time"

	"driftcheck/internal/alert"
	"driftcheck/internal/config"
	"driftcheck/internal/drift"
	"driftcheck/internal/history"
	"driftcheck/internal/schedule"
	"driftcheck/internal/snapshot"
)

// stubFetcher is a snapshot.Fetcher that returns a fixed map.
type stubFetcher struct {
	values map[string]string
	err    error
}

func (s *stubFetcher) Fetch(serviceName string) (map[string]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.values, nil
}

func newRunnerConfig(t *testing.T, dir string) schedule.RunnerConfig {
	t.Helper()

	cfg := &config.Config{
		ServiceName: "svc",
		Expected:    map[string]string{"KEY": "val"},
	}

	fetcher := &stubFetcher{values: map[string]string{"KEY": "val"}}
	det := drift.NewDetector()

	var buf bytes.Buffer
	rep := drift.NewReporter(&buf)

	store, err := history.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	var logBuf bytes.Buffer
	logger := log.New(&logBuf, "", 0)
	notifier := alert.NewLogNotifier(logger)
	disp := alert.NewDispatcher(notifier)

	return schedule.RunnerConfig{
		Configs:    []*config.Config{cfg},
		Fetcher:    fetcher,
		Detector:   det,
		Reporter:   rep,
		Store:      store,
		Dispatcher: disp,
		Logger:     log.New(&logBuf, "", 0),
	}
}

func TestRunner_RunOnce_NoDrift(t *testing.T) {
	dir := t.TempDir()
	cfg := newRunnerConfig(t, dir)

	runner := schedule.NewRunner(cfg)

	ctx := context.Background()
	err := runner.RunOnce(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestRunner_RunOnce_FetchError(t *testing.T) {
	dir := t.TempDir()
	cfg := newRunnerConfig(t, dir)

	// Replace fetcher with one that errors.
	cfg.Fetcher = &stubFetcher{err: errors.New("env unavailable")}

	runner := schedule.NewRunner(cfg)

	ctx := context.Background()
	err := runner.RunOnce(ctx)
	if err == nil {
		t.Fatal("expected error from fetch failure, got nil")
	}
	if !strings.Contains(err.Error(), "env unavailable") {
		t.Errorf("unexpected error message: %v", err)
	}
}

func TestRunner_RunOnce_WithDrift(t *testing.T) {
	dir := t.TempDir()
	cfg := newRunnerConfig(t, dir)

	// Fetcher returns a value that doesn't match the declared config.
	cfg.Fetcher = &stubFetcher{values: map[string]string{"KEY": "wrong"}}

	var buf bytes.Buffer
	cfg.Reporter = drift.NewReporter(&buf)

	runner := schedule.NewRunner(cfg)

	ctx := context.Background()
	err := runner.RunOnce(ctx)
	if err != nil {
		t.Fatalf("RunOnce should not return error on drift, got %v", err)
	}

	if !strings.Contains(buf.String(), "KEY") {
		t.Errorf("expected drift report to mention KEY, got: %s", buf.String())
	}
}

func TestRunner_RunOnce_PersistsHistory(t *testing.T) {
	dir := t.TempDir()
	cfg := newRunnerConfig(t, dir)

	runner := schedule.NewRunner(cfg)

	ctx := context.Background()
	if err := runner.RunOnce(ctx); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	store, err := history.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	records, err := store.List("svc")
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 history record, got %d", len(records))
	}
}

func TestRunner_RunOnce_ContextCancelled(t *testing.T) {
	dir := t.TempDir()
	cfg := newRunnerConfig(t, dir)

	runner := schedule.NewRunner(cfg)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// Wait for context to expire.
	time.Sleep(5 * time.Millisecond)

	err := runner.RunOnce(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}

func TestRunner_BuildsSnapshotCorrectly(t *testing.T) {
	dir := t.TempDir()
	cfg := newRunnerConfig(t, dir)

	expected := map[string]string{"APP_ENV": "production", "PORT": "8080"}
	cfg.Configs = []*config.Config{
		{ServiceName: "web", Expected: expected},
	}
	cfg.Fetcher = &stubFetcher{values: expected}

	runner := schedule.NewRunner(cfg)

	ctx := context.Background()
	if err := runner.RunOnce(ctx); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	// Verify snapshot values are stored in history.
	store, _ := history.NewStore(dir)
	records, _ := store.List("web")
	if len(records) == 0 {
		t.Fatal("no records found for service 'web'")
	}

	snap := records[0].Snapshot
	for k, v := range expected {
		got, err := snap.Get(k)
		if err != nil {
			t.Errorf("snapshot missing key %q: %v", k, err)
			continue
		}
		if got != v {
			t.Errorf("snapshot[%q] = %q, want %q", k, got, v)
		}
	}
	_ = snapshot.New // ensure import is used
}
