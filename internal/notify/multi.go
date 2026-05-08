package notify

import (
	"errors"
	"fmt"

	"driftcheck/internal/alert"
)

// Notifier is the common interface for all notification backends.
type Notifier interface {
	Notify(result alert.Result) error
}

// MultiNotifier fans a single alert out to multiple Notifier implementations.
type MultiNotifier struct {
	notifiers []Notifier
	stopOnErr bool
}

// NewMultiNotifier creates a MultiNotifier that calls each notifier in order.
// If stopOnErr is true the first failure aborts remaining notifiers.
func NewMultiNotifier(stopOnErr bool, notifiers ...Notifier) *MultiNotifier {
	return &MultiNotifier{
		notifiers: notifiers,
		stopOnErr: stopOnErr,
	}
}

// Notify dispatches the result to every registered notifier.
func (m *MultiNotifier) Notify(result alert.Result) error {
	var errs []error
	for _, n := range m.notifiers {
		if err := n.Notify(result); err != nil {
			if m.stopOnErr {
				return fmt.Errorf("multi notifier: %w", err)
			}
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
