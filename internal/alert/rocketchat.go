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

// rocketChatColor returns an attachment color string for the given alert level.
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

// NewRocketChatNotifier creates a new RocketChatNotifier.
func NewRocketChatNotifier(webhookURL string) (*RocketChatNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("rocketchat: webhook URL must not be empty")
	}
	return &RocketChatNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}, nil
}

// Send delivers the alert as a Rocket.Chat message attachment.
func (n *RocketChatNotifier) Send(a Alert) error {
	type attachment struct {
		Title string `json:"title"`
		Text  string `json:"text"`
		Color string `json:"color"`
	}
	type payload struct {
		Text        string       `json:"text"`
		Attachments []attachment `json:"attachments"`
	}
	p := payload{
		Text: "VaultWatch Secret Expiration Alert",
		Attachments: []attachment{
			{
				Title: fmt.Sprintf("[%s] %s", a.Level, a.SecretPath),
				Text:  fmt.Sprintf("Expires in: %s\n%s", a.TimeLeft.Round(0), a.Message),
				Color: rocketChatColor(a.Level),
			},
		},
	}
	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("rocketchat: marshal payload: %w", err)
	}
	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("rocketchat: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("rocketchat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
