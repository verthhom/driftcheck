// Package schedule provides a simple periodic scheduler for running
// drift-check jobs at a configurable interval.
//
// Usage:
//
//	scheduler := schedule.NewScheduler(
//		5*time.Minute,
//		func(ctx context.Context) error {
//			// run drift checks here
//			return nil
//		},
//		nil,
//	)
//	scheduler.Run(ctx)
//
// The job is executed immediately on the first call and then on every
// subsequent tick. If the context is cancelled the scheduler stops
// cleanly after the current tick completes.
package schedule
