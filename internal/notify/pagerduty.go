package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"driftcheck/internal/drift"
)

const defaultPagerDutyEndpoint = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends drift alerts to PagerDuty via the Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	endpoint       string
	client         *http.Client
}

type pdPayload struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdDetails `json:"payload"`
}

type pdDetails struct {
	Summary   string `json:"summary"`
	Source    string `json:"source"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
}

// NewPagerDutyNotifier creates a PagerDutyNotifier using the given integration key.
func NewPagerDutyNotifier(integrationKey string) *PagerDutyNotifier {
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		endpoint:       defaultPagerDutyEndpoint,
		client:         &http.Client{Timeout: 10 * time.Second},
	}
}

// Notify sends a PagerDuty trigger event when drift is detected.
func (p *PagerDutyNotifier) Notify(result drift.Result) error {
	if p.integrationKey == "" {
		return fmt.Errorf("pagerduty: integration key must not be empty")
	}
	if !result.HasDrift() {
		return nil
	}

	summary := fmt.Sprintf("Drift detected in service %q: %d key(s) differ",
		result.ServiceName, len(result.Drifted))

	body := pdPayload{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdDetails{
			Summary:   summary,
			Source:    "driftcheck",
			Severity:  "error",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	resp, err := p.client.Post(p.endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: send event: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
