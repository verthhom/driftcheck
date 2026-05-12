package notify

// SetEndpoint overrides the PagerDuty events endpoint for testing.
func (p *PagerDutyNotifier) SetEndpoint(url string) {
	p.endpoint = url
}
