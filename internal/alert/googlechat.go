package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GoogleChatNotifier sends alerts to a Google Chat webhook.
type GoogleChatNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewGoogleChatNotifier creates a new GoogleChatNotifier.
// webhookURL must be a valid Google Chat incoming webhook URL.
func NewGoogleChatNotifier(webhookURL string) (*GoogleChatNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("googlechat: webhook URL must not be empty")
	}
	return &GoogleChatNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}, nil
}

type googleChatPayload struct {
	Text string `json:"text"`
}

// Send delivers the alert to Google Chat via the configured webhook.
func (n *GoogleChatNotifier) Send(a Alert) error {
	msg := fmt.Sprintf("*[%s] VaultWatch Alert*\n%s", a.Level, a.Message)
	payload := googleChatPayload{Text: msg}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("googlechat: failed to marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("googlechat: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("googlechat: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
