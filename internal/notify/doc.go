// Package notify provides notifier implementations for delivering drift
// alerts through external channels.
//
// Available notifiers:
//
//   - WebhookNotifier – HTTP POST to an arbitrary endpoint.
//   - EmailNotifier   – SMTP email delivery.
//   - SlackNotifier   – Slack incoming-webhook message.
//   - MultiNotifier   – Fan-out across multiple notifiers.
//
// All notifiers implement a common Notify(ctx, drift.Result) error interface
// so they can be composed freely with the alert.Dispatcher.
package notify
