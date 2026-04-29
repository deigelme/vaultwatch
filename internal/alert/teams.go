package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// teamsPayload is the Adaptive Card payload for Microsoft Teams via Incoming Webhook.
type teamsPayload struct {
	Type       string         `json:"type"`
	Attachments []teamsAttach `json:"attachments"`
}

type teamsAttach struct {
	ContentType string       `json:"contentType"`
	Content     teamsContent `json:"content"`
}

type teamsContent struct {
	Schema  string      `json:"$schema"`
	Type    string      `json:"type"`
	Version string      `json:"version"`
	Body    []teamsBody `json:"body"`
}

type teamsBody struct {
	Type   string `json:"type"`
	Text   string `json:"text"`
	Weight string `json:"weight,omitempty"`
	Size   string `json:"size,omitempty"`
}

// TeamsNotifier sends alerts to a Microsoft Teams channel via Incoming Webhook.
type TeamsNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewTeamsNotifier creates a TeamsNotifier. Returns an error if webhookURL is empty.
func NewTeamsNotifier(webhookURL string) (*TeamsNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("teams webhook URL must not be empty")
	}
	return &TeamsNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send posts the alert to the configured Teams webhook.
func (t *TeamsNotifier) Send(a Alert) error {
	payload := teamsPayload{
		Type: "message",
		Attachments: []teamsAttach{
			{
				ContentType: "application/vnd.microsoft.card.adaptive",
				Content: teamsContent{
					Schema:  "http://adaptivecards.io/schemas/adaptive-card.json",
					Type:    "AdaptiveCard",
					Version: "1.4",
					Body: []teamsBody{
						{Type: "TextBlock", Text: "VaultWatch Alert", Weight: "Bolder", Size: "Medium"},
						{Type: "TextBlock", Text: a.String()},
					},
				},
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("teams: marshal payload: %w", err)
	}

	resp, err := t.client.Post(t.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("teams: http post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("teams: unexpected status %d", resp.StatusCode)
	}
	return nil
}
