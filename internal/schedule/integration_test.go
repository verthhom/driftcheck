package schedule_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"driftcheck/internal/schedule"
)

// TestScheduler_CollectsResults verifies that results produced by the job
// inside a goroutine are visible to the test after the scheduler stops.
func TestScheduler_CollectsResults(t *testing.T) {
	var mu sync.Mutex
	var results []int

	counter := 0
	job := func(_ context.Context) error {
		mu.Lock()
		results = append(results, counter)
		counter++
		mu.Unlock()
		return nil
	}

	s := schedule.NewScheduler(15*time.Millisecond, job, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	defer cancel()

	s.Run(ctx)

	mu.Lock()
	defer mu.Unlock()

	if len(results) < 2 {
		t.Fatalf("expected at least 2 collected results, got %d", len(results))
	}

	for i, v := range results {
		if v != i {
			t.Fatalf("results[%d] = %d, want %d", i, v, i)
		}
	}
}

// TestScheduler_NilLoggerDefaults ensures passing a nil logger does not panic.
func TestScheduler_NilLoggerDefaults(t *testing.T) {
	job := func(_ context.Context) error { return nil }

	s := schedule.NewScheduler(1*time.Second, job, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	// Should not panic.
	s.Run(ctx)
}
