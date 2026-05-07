// Package notify provides outbound notification integrations for drift events.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookPayload is the JSON body sent to a webhook endpoint.
type WebhookPayload struct {
	Service   string    `json:"service"`
	Drifted   bool      `json:"drifted"`
	Keys      []string  `json:"drifted_keys,omitempty"`
	Severity  string    `json:"severity"`
	Timestamp time.Time `json:"timestamp"`
}

// WebhookNotifier sends drift alerts to an HTTP endpoint.
type WebhookNotifier struct {
	endpoint string
	client   *http.Client
}

// NewWebhookNotifier returns a WebhookNotifier that posts to the given endpoint.
// A zero timeout defaults to 10 seconds.
func NewWebhookNotifier(endpoint string, timeout time.Duration) *WebhookNotifier {
	if timeout == 0 {
		timeout = 10 * time.Second
	}
	return &WebhookNotifier{
		endpoint: endpoint,
		client:   &http.Client{Timeout: timeout},
	}
}

// Notify serialises payload and POSTs it to the configured endpoint.
func (w *WebhookNotifier) Notify(ctx context.Context, p WebhookPayload) error {
	if w.endpoint == "" {
		return fmt.Errorf("webhook: endpoint must not be empty")
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status %d from %s", resp.StatusCode, w.endpoint)
	}
	return nil
}
