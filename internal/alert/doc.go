// Package alert implements drift alerting and notification for driftcheck.
//
// It provides the Alert type which captures details about a detected drift
// event, the Notifier interface for pluggable notification backends, and the
// Dispatcher which bridges drift.Result values to one or more Notifier
// implementations.
//
// # Usage
//
//	notifier := alert.NewLogNotifier(os.Stderr)
//	dispatcher := alert.NewDispatcher(notifier)
//	if err := dispatcher.Dispatch(result); err != nil {
//		log.Fatal(err)
//	}
package alert
