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

// NewGoogleChatNotifier returns a GoogleChatNotifier or an error if webhookURL is empty.
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

// Send delivers the alert to the configured Google Chat webhook.
func (n *GoogleChatNotifier) Send(a Alert) error {
	payload := googleChatPayload{
		Text: fmt.Sprintf("*[%s] VaultWatch Alert*\n%s", a.Level, a.String()),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("googlechat: marshal payload: %w", err)
	}
	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("googlechat: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("googlechat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
