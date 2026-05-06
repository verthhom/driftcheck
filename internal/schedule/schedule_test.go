package schedule_test

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"driftcheck/internal/schedule"
)

func TestScheduler_RunsJobImmediately(t *testing.T) {
	var count atomic.Int32

	job := func(_ context.Context) error {
		count.Add(1)
		return nil
	}

	s := schedule.NewScheduler(10*time.Second, job, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	s.Run(ctx)

	if count.Load() < 1 {
		t.Fatal("expected job to run at least once immediately")
	}
}

func TestScheduler_RunsJobMultipleTimes(t *testing.T) {
	var count atomic.Int32

	job := func(_ context.Context) error {
		count.Add(1)
		return nil
	}

	s := schedule.NewScheduler(20*time.Millisecond, job, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Millisecond)
	defer cancel()

	s.Run(ctx)

	if got := count.Load(); got < 3 {
		t.Fatalf("expected at least 3 executions, got %d", got)
	}
}

func TestScheduler_StopsOnContextCancel(t *testing.T) {
	var count atomic.Int32

	job := func(_ context.Context) error {
		count.Add(1)
		return nil
	}

	s := schedule.NewScheduler(5*time.Millisecond, job, nil)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		s.Run(ctx)
		close(done)
	}()

	time.Sleep(30 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// ok
	case <-time.After(200 * time.Millisecond):
		t.Fatal("scheduler did not stop after context cancellation")
	}
}

func TestScheduler_JobErrorDoesNotStop(t *testing.T) {
	var count atomic.Int32
	errJob := errors.New("boom")

	job := func(_ context.Context) error {
		count.Add(1)
		return errJob
	}

	s := schedule.NewScheduler(20*time.Millisecond, job, nil)
	ctx, cancel := context.WithTimeout(context.Background(), 70*time.Millisecond)
	defer cancel()

	s.Run(ctx)

	if got := count.Load(); got < 2 {
		t.Fatalf("expected scheduler to keep running after error, got %d executions", got)
	}
}
