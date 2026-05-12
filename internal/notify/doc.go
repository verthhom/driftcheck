// Package notify provides notification backends for delivering drift alerts
// to external systems.
//
// Supported notifiers:
//
//   - WebhookNotifier  – HTTP POST to an arbitrary endpoint
//   - SlackNotifier    – Slack Incoming Webhooks
//   - EmailNotifier    – SMTP email delivery
//   - PagerDutyNotifier – PagerDuty Events API v2
//
// All notifiers implement the alert.Notifier interface and can be composed
// with MultiNotifier to fan out alerts to several destinations simultaneously.
//
// Notifiers are silent (return nil) when the supplied drift.Result contains
// no drifted keys, so callers do not need to gate on drift presence.
package notify
