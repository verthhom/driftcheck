// Package notify provides Notifier implementations that deliver drift results
// to external systems such as Slack, PagerDuty, email, and generic webhooks.
//
// Decorator notifiers (MultiNotifier, FilteredNotifier, DedupeNotifier,
// RateLimitedNotifier, RetryNotifier, ThrottleNotifier) wrap any Notifier and
// add cross-cutting behaviour without modifying the underlying transport.
//
// ThrottleNotifier differs from RateLimitedNotifier in that it enforces a
// per-service silence window: once a notification has been forwarded for a
// given service, further calls within the window are suppressed rather than
// queued. This prevents alert storms when a service drifts repeatedly within a
// short period.
package notify
