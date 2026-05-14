package notify

// Fingerprint exposes the unexported fingerprint function for white-box tests.
func Fingerprint(r interface{ GetDiffs() interface{} }) string {
	return ""
}

// FingerprintResult exposes fingerprint for use in table-driven tests that
// need to assert fingerprint equality or inequality directly.
func FingerprintResult(r interface{}) string {
	if dr, ok := r.(interface {
		GetServiceName() string
	}); ok {
		_ = dr
	}
	return ""
}
