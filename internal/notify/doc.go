// Package notify provides a collection of Notifier implementations and
// decorators for delivering drift-check results to external systems.
//
// Core notifiers
//
//   - WebhookNotifier  – HTTP POST to an arbitrary endpoint.
//   - EmailNotifier    – SMTP delivery.
//   - SlackNotifier    – Slack incoming webhook.
//   - PagerDutyNotifier – PagerDuty Events API v2.
//
// Decorators (wrap any Notifier)
//
//   - MultiNotifier        – fan-out to multiple notifiers.
//   - FilteredNotifier     – conditional forwarding.
//   - DedupeNotifier       – suppress duplicate alerts.
//   - ThrottleNotifier     – sliding-window rate limit.
//   - RateLimitedNotifier  – cooldown-based rate limit.
//   - RetryNotifier        – automatic retry with back-off.
//   - CircuitBreakerNotifier – open/half-open/closed circuit.
//   - BatchNotifier        – accumulate then flush.
//   - BufferedNotifier     – async channel-backed buffer.
//   - WindowNotifier       – rolling-window deduplication per service.
//   - AuditNotifier        – structured audit logging.
//   - TransformingNotifier – mutate results before forwarding.
//   - FallbackNotifier     – primary + ordered fallback chain.
//   - DigestNotifier       – periodic digest delivery.
//   - RoutingNotifier      – rule-based routing.
//   - SamplingNotifier     – probabilistic sampling.
//   - ReplayNotifier       – record and replay results.
//   - TimeoutNotifier      – per-call deadline enforcement.
//   - PriorityNotifier     – high/low priority routing.
//   - EnvelopeNotifier     – metadata stamping.
//   - ShadowNotifier       – shadow traffic to secondary.
//   - SquashNotifier       – collapse rapid repeated calls.
//   - HealthCheckNotifier  – liveness tracking.
//   - TaggingNotifier      – annotation injection.
//   - SpoolNotifier        – disk-backed spooling.
//   - FallbackNotifier     – fallback chain.
package notify
