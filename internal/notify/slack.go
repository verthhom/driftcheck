package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"driftcheck/internal/drift"
)

// SlackNotifier sends drift alerts to a Slack webhook URL.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewSlackNotifier creates a SlackNotifier that posts to the given Slack
// incoming-webhook URL. A zero timeout defaults to 10 seconds.
func NewSlackNotifier(webhookURL string, timeout time.Duration) *SlackNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &SlackNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: timeout},
	}
}

type slackPayload struct {
	Text string `json:"text"`
}

// Notify sends a Slack message summarising the drift result.
func (s *SlackNotifier) Notify(ctx context.Context, result drift.Result) error {
	if s.webhookURL == "" {
		return fmt.Errorf("slack: webhook URL must not be empty")
	}

	var msg string
	if !result.HasDrift() {
		msg = fmt.Sprintf(":white_check_mark: *%s* — no drift detected", result.ServiceName)
	} else {
		msg = fmt.Sprintf(":warning: *%s* — drift detected in %d key(s)", result.ServiceName, len(result.Drifted))
		for _, d := range result.Drifted {
			msg += fmt.Sprintf("\n  • `%s`: expected `%s`, got `%s`", d.Key, d.Expected, d.Actual)
		}
	}

	body, err := json.Marshal(slackPayload{Text: msg})
	if err != nil {
		return fmt.Errorf("slack: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("slack: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("slack: unexpected status %d", resp.StatusCode)
	}
	return nil
}
