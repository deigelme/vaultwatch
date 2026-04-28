package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SlackNotifier sends alert notifications to a Slack webhook.
type SlackNotifier struct {
	webhookURL string
	client     *http.Client
}

type slackPayload struct {
	Text string `json:"text"`
}

// NewSlackNotifier creates a new SlackNotifier with the given webhook URL.
func NewSlackNotifier(webhookURL string) (*SlackNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("slack webhook URL must not be empty")
	}
	return &SlackNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

// Send dispatches the alert to the configured Slack webhook.
func (s *SlackNotifier) Send(a Alert) error {
	payload := slackPayload{
		Text: fmt.Sprintf("[%s] %s", a.Level, a.String()),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("slack: failed to marshal payload: %w", err)
	}

	resp, err := s.client.Post(s.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("slack: failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("slack: unexpected status code: %d", resp.StatusCode)
	}
	return nil
}
