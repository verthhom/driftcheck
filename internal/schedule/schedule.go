// Package schedule provides periodic drift-check scheduling.
package schedule

import (
	"context"
	"log"
	"time"
)

// Job is a function executed on each tick.
type Job func(ctx context.Context) error

// Scheduler runs a Job at a fixed interval until the context is cancelled.
type Scheduler struct {
	interval time.Duration
	job      Job
	logger   *log.Logger
}

// NewScheduler creates a Scheduler that runs job every interval.
// If logger is nil it defaults to the standard logger.
func NewScheduler(interval time.Duration, job Job, logger *log.Logger) *Scheduler {
	if logger == nil {
		logger = log.Default()
	}
	return &Scheduler{
		interval: interval,
		job:      job,
		logger:   logger,
	}
}

// Run blocks until ctx is cancelled, executing the job on every tick.
// The job is also executed immediately on the first call before waiting.
func (s *Scheduler) Run(ctx context.Context) {
	s.tick(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.tick(ctx)
		case <-ctx.Done():
			s.logger.Println("schedule: stopped")
			return
		}
	}
}

func (s *Scheduler) tick(ctx context.Context) {
	if err := s.job(ctx); err != nil {
		s.logger.Printf("schedule: job error: %v", err)
	}
}
