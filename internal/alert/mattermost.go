package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// MattermostNotifier sends alerts to a Mattermost incoming webhook.
type MattermostNotifier struct {
	webhookURL string
	channel    string
	username   string
	client     *http.Client
}

type mattermostPayload struct {
	Channel  string `json:"channel,omitempty"`
	Username string `json:"username,omitempty"`
	Text     string `json:"text"`
	IconEmoji string `json:"icon_emoji,omitempty"`
}

// NewMattermostNotifier creates a MattermostNotifier.
// webhookURL must be a valid Mattermost incoming webhook URL.
func NewMattermostNotifier(webhookURL, channel, username string) (*MattermostNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("mattermost: webhook URL must not be empty")
	}
	return &MattermostNotifier{
		webhookURL: webhookURL,
		channel:    channel,
		username:   username,
		client:     &http.Client{},
	}, nil
}

// Send posts an alert message to Mattermost.
func (m *MattermostNotifier) Send(a Alert) error {
	emoji := ":warning:"
	if a.Level == Critical {
		emoji = ":rotating_light:"
	}

	payload := mattermostPayload{
		Channel:   m.channel,
		Username:  m.username,
		Text:      fmt.Sprintf("%s %s", emoji, a.String()),
		IconEmoji: emoji,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("mattermost: failed to marshal payload: %w", err)
	}

	resp, err := m.client.Post(m.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("mattermost: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mattermost: unexpected status %d", resp.StatusCode)
	}
	return nil
}
