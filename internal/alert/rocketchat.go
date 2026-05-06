package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// RocketChatNotifier sends alerts to a Rocket.Chat incoming webhook.
type RocketChatNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewRocketChatNotifier creates a new RocketChatNotifier.
// webhookURL must be a valid Rocket.Chat incoming webhook URL.
func NewRocketChatNotifier(webhookURL string) (*RocketChatNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("rocketchat: webhook URL must not be empty")
	}
	return &RocketChatNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}, nil
}

type rocketChatPayload struct {
	Text        string             `json:"text"`
	Attachments []rocketAttachment `json:"attachments,omitempty"`
}

type rocketAttachment struct {
	Title  string `json:"title"`
	Text   string `json:"text"`
	Color  string `json:"color"`
}

func rocketChatColor(level Level) string {
	switch level {
	case LevelCritical:
		return "#FF0000"
	case LevelWarning:
		return "#FFA500"
	default:
		return "#36A64F"
	}
}

// Send delivers an alert to the configured Rocket.Chat webhook.
func (n *RocketChatNotifier) Send(a Alert) error {
	payload := rocketChatPayload{
		Text: fmt.Sprintf("*VaultWatch Alert* [%s]", a.Level),
		Attachments: []rocketAttachment{
			{
				Title: a.SecretPath,
				Text:  a.String(),
				Color: rocketChatColor(a.Level),
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("rocketchat: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("rocketchat: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("rocketchat: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
