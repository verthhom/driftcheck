// Package notify provides pluggable outbound notification backends for
// driftcheck alerts.
//
// Currently supported backends:
//
//   - WebhookNotifier — POSTs a JSON payload to an HTTP endpoint whenever
//     drift is detected. The payload includes the service name, the list of
//     drifted keys, a severity label, and the detection timestamp.
//
// Typical usage:
//
//	n := notify.NewWebhookNotifier("https://hooks.example.com/drift", 0)
//	err := n.Notify(ctx, notify.WebhookPayload{
//		Service:  "api-gateway",
//		Drifted:  true,
//		Keys:     []string{"PORT", "LOG_LEVEL"},
//		Severity: "high",
//		Timestamp: time.Now(),
//	})
package notify
