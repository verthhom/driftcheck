// Package notify provides Notifier implementations for delivering drift
// detection results to external systems.
//
// Core notifiers (Webhook, Slack, Email, PagerDuty) send results to a single
// destination. Decorator notifiers wrap a core notifier to add cross-cutting
// behaviour:
//
//   - MultiNotifier   – fan-out to multiple notifiers
//   - FilteredNotifier – gate forwarding behind a predicate
//   - RoutingNotifier  – route to the first matching rule
//   - DedupeNotifier   – suppress duplicate alerts
//   - ThrottleNotifier – limit alert frequency per window
//   - RateLimitedNotifier – enforce a minimum cooldown between sends
//   - RetryNotifier    – retry on transient failures
//   - CircuitBreakerNotifier – open circuit after repeated failures
//   - BatchNotifier    – coalesce results over a time window
//   - BufferedNotifier – queue results and flush asynchronously
//   - DigestNotifier   – accumulate and flush combined digests
//   - TransformingNotifier – mutate results before forwarding
//   - FallbackNotifier – try primary, fall back on error
//   - AuditNotifier    – log every notification attempt
package notify
