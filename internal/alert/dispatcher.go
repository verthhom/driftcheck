package alert

import (
	"fmt"
	"log"

	"driftcheck/internal/drift"
)

// Notifier is implemented by any type that can receive a drift result.
type Notifier interface {
	Notify(result drift.Result) error
}

// Dispatcher fans a drift result out to one or more Notifier implementations.
type Dispatcher struct {
	notifiers []Notifier
	logger    *log.Logger
}

// NewDispatcher creates a Dispatcher that will call each supplied Notifier.
func NewDispatcher(logger *log.Logger, notifiers ...Notifier) *Dispatcher {
	if logger == nil {
		logger = log.Default()
	}
	return &Dispatcher{notifiers: notifiers, logger: logger}
}

// Dispatch sends result to every registered Notifier.
// All notifiers are attempted; a combined error is returned if any fail.
func (d *Dispatcher) Dispatch(result drift.Result) error {
	var errs []error
	for _, n := range d.notifiers {
		if err := n.Notify(result); err != nil {
			d.logger.Printf("alert dispatcher: notifier error: %v", err)
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("alert dispatcher: %d notifier(s) failed: %v", len(errs), errs)
	}
	return nil
}

// driftedKeys returns the keys that have drifted in result.
func driftedKeys(result drift.Result) []string {
	keys := make([]string, 0, len(result.Drifted))
	for k := range result.Drifted {
		keys = append(keys, k)
	}
	return keys
}
