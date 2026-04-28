package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// WebhookNotifier sends alert notifications to a generic HTTP webhook endpoint.
type WebhookNotifier struct {
	url    string
	client *http.Client
}

type webhookPayload struct {
	Level     string    `json:"level"`
	Secret    string    `json:"secret"`
	ExpiresIn string    `json:"expires_in"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

// NewWebhookNotifier creates a WebhookNotifier that posts JSON payloads to url.
func NewWebhookNotifier(url string) (*WebhookNotifier, error) {
	if url == "" {
		return nil, fmt.Errorf("webhook url must not be empty")
	}
	return &WebhookNotifier{
		url: url,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers the alert to the configured webhook URL.
func (w *WebhookNotifier) Send(a Alert) error {
	payload := webhookPayload{
		Level:     string(a.Level),
		Secret:    a.SecretPath,
		ExpiresIn: a.TimeLeft.String(),
		Message:   a.String(),
		Timestamp: time.Now().UTC(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: failed to marshal payload: %w", err)
	}

	resp, err := w.client.Post(w.url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
