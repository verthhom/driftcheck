// Package notify provides notification backends for driftcheck.
//
// Supported notifiers:
//   - WebhookNotifier: HTTP POST to an arbitrary endpoint.
//   - SlackNotifier: Slack incoming webhook.
//   - EmailNotifier: SMTP email delivery.
//   - PagerDutyNotifier: PagerDuty Events API v2.
//   - MultiNotifier: fan-out to multiple notifiers.
//   - RateLimitedNotifier: suppresses repeated alerts within a cooldown window.
//   - RetryNotifier: retries a notifier on transient failure with configurable
//     backoff and maximum attempt count.
//
// All notifiers implement the drift.Notifier interface and accept a
// context.Context so callers can enforce deadlines and cancellation.
package notify
