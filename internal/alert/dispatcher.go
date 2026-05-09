package alert

import (
	"fmt"
	"time"

	"github.com/example/driftcheck/internal/drift"
)

// Dispatcher evaluates drift results and dispatches alerts via registered notifiers.
type Dispatcher struct {
	notifiers []Notifier
}

// NewDispatcher creates a Dispatcher with the provided notifiers.
func NewDispatcher(notifiers ...Notifier) *Dispatcher {
	return &Dispatcher{notifiers: notifiers}
}

// Dispatch converts a drift.Result into an Alert and sends it to all notifiers.
// It returns the first error encountered, if any.
func (d *Dispatcher) Dispatch(result drift.Result) error {
	keys := driftedKeys(result)
	sev := SeverityFor(len(keys))

	msg := fmt.Sprintf("no drift detected for service %q", result.ServiceName)
	if len(keys) > 0 {
		msg = fmt.Sprintf("drift detected for service %q: %d key(s) differ", result.ServiceName, len(keys))
	}

	a := Alert{
		ServiceName: result.ServiceName,
		Severity:    sev,
		Message:     msg,
		DriftKeys:   keys,
		OccurredAt:  time.Now(),
	}

	for _, n := range d.notifiers {
		if err := n.Notify(a); err != nil {
			return fmt.Errorf("alert dispatcher: notifier %T failed: %w", n, err)
		}
	}
	return nil
}

// DispatchAll calls Dispatch for each result and collects all errors.
// Unlike Dispatch, it does not stop on the first error.
func (d *Dispatcher) DispatchAll(results []drift.Result) []error {
	var errs []error
	for _, r := range results {
		if err := d.Dispatch(r); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

// driftedKeys extracts the list of keys that have drift from a Result.
func driftedKeys(result drift.Result) []string {
	var keys []string
	for _, d := range result.Diffs {
		keys = append(keys, d.Key)
	}
	return keys
}
