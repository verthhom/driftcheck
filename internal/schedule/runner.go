package schedule

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"driftcheck/internal/alert"
	"driftcheck/internal/config"
	"driftcheck/internal/drift"
	"driftcheck/internal/history"
	"driftcheck/internal/snapshot"
)

// RunOptions holds the dependencies and configuration needed for a scheduled
// drift-check run.
type RunOptions struct {
	ConfigDir   string
	HistoryDir  string
	Notifiers   []alert.Notifier
	Logger      *log.Logger
}

// Runner orchestrates a single end-to-end drift-check cycle: it loads all
// service configs, fetches live snapshots, checks for drift, persists the
// results to history, and dispatches alerts when drift is detected.
type Runner struct {
	opts   RunOptions
	logger *log.Logger
}

// NewRunner creates a Runner with the supplied options. If opts.Logger is nil
// it defaults to writing to os.Stderr.
func NewRunner(opts RunOptions) *Runner {
	if opts.Logger == nil {
		opts.Logger = log.New(os.Stderr, "[driftcheck] ", log.LstdFlags)
	}
	return &Runner{opts: opts, logger: opts.Logger}
}

// Run performs one full drift-check cycle. It returns the first error
// encountered, but always attempts to process every service config before
// returning.
func (r *Runner) Run(ctx context.Context) error {
	loader := config.NewLoader()
	configs, err := loader.LoadAll(r.opts.ConfigDir)
	if err != nil {
		return fmt.Errorf("loading configs from %s: %w", r.opts.ConfigDir, err)
	}

	store, err := history.NewStore(r.opts.HistoryDir)
	if err != nil {
		return fmt.Errorf("opening history store at %s: %w", r.opts.HistoryDir, err)
	}

	dispatcher := alert.NewDispatcher(r.opts.Notifiers...)
	detector := drift.NewDetector()
	var firstErr error

	for _, cfg := range configs {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if err := r.processService(ctx, cfg, detector, store, dispatcher); err != nil {
			r.logger.Printf("error processing service %q: %v", cfg.ServiceName, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

// processService runs the full drift-check pipeline for a single service.
func (r *Runner) processService(
	ctx context.Context,
	cfg *config.Config,
	detector *drift.Detector,
	store *history.Store,
	dispatcher *alert.Dispatcher,
) error {
	fetcher := snapshot.NewEnvFetcher(cfg.ServiceName)
	snap, err := snapshot.New(cfg, fetcher)
	if err != nil {
		return fmt.Errorf("snapshot: %w", err)
	}

	result := detector.Check(snap)

	record := history.Record{
		ServiceName: cfg.ServiceName,
		CheckedAt:   time.Now().UTC(),
		Result:      result,
	}
	if err := store.Save(record); err != nil {
		return fmt.Errorf("saving history: %w", err)
	}

	if err := dispatcher.Dispatch(ctx, result); err != nil {
		return fmt.Errorf("dispatching alerts: %w", err)
	}

	if result.HasDrift() {
		r.logger.Printf("drift detected for service %q (%d key(s))",
			cfg.ServiceName, len(result.Diffs))
	} else {
		r.logger.Printf("no drift for service %q", cfg.ServiceName)
	}

	return nil
}

// WriteSummary writes a human-readable summary of recent history records for
// all services to w. It is a convenience wrapper around history.Summarise.
func (r *Runner) WriteSummary(w io.Writer) error {
	store, err := history.NewStore(r.opts.HistoryDir)
	if err != nil {
		return fmt.Errorf("opening history store: %w", err)
	}

	records, err := store.List("")
	if err != nil {
		return fmt.Errorf("listing history: %w", err)
	}

	summary := history.Summarise(records)
	for _, s := range summary {
		fmt.Fprintf(w, "service=%-20s checks=%d drifted=%d last_drift=%s\n",
			s.ServiceName, s.TotalChecks, s.DriftedChecks, formatTime(s.LastDriftAt))
	}
	return nil
}

func formatTime(t *time.Time) string {
	if t == nil {
		return "never"
	}
	return t.Format(time.RFC3339)
}
