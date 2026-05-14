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
		return nil, fmt.Errorf("googlechat: webhookURL must not be empty")
	}
	return &GoogleChatNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}, nil
}

// Send delivers an Alert to Google Chat.
func (n *GoogleChatNotifier) Send(a Alert) error {
	payload := map[string]string{
		"text": fmt.Sprintf("*[%s] VaultWatch Alert*\n%s", a.Level, a.String()),
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("googlechat: marshal payload: %w", err)
	}
	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(b))
	if err != nil {
		return fmt.Errorf("googlechat: post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("googlechat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
