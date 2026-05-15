package notify

// AuditNotifierName exposes the configured notifier name for testing.
func AuditNotifierName(a *AuditNotifier) string {
	return a.name
}
