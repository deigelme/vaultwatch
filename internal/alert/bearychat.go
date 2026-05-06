package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// BearyChat notifier sends alerts to a BearyChat incoming webhook.
type bearyChatNotifier struct {
	webhookURL string
	client     *http.Client
}

type bearyChatPayload struct {
	Text        string                   `json:"text"`
	Markdown    bool                     `json:"markdown"`
	Attachments []bearyChatAttachment    `json:"attachments,omitempty"`
}

type bearyChatAttachment struct {
	Text  string `json:"text"`
	Color string `json:"color"`
}

// NewBearyChatNotifier creates a new BearyChat notifier.
// webhookURL must be non-empty.
func NewBearyChatNotifier(webhookURL string) (Notifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("bearychat: webhook URL must not be empty")
	}
	return &bearyChatNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}, nil
}

func (n *bearyChatNotifier) Send(a Alert) error {
	color := "#36a64f"
	switch a.Level {
	case LevelWarning:
		color = "#ffae42"
	case LevelCritical:
		color = "#e03e2f"
	}

	payload := bearyChatPayload{
		Text:     fmt.Sprintf("**VaultWatch Alert** — %s", a.SecretPath),
		Markdown: true,
		Attachments: []bearyChatAttachment{
			{
				Text:  a.String(),
				Color: color,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("bearychat: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("bearychat: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bearychat: unexpected status %d", resp.StatusCode)
	}
	return nil
}
