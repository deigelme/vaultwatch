package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// LarkNotifier sends alerts to a Lark (Feishu) incoming webhook.
type LarkNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewLarkNotifier creates a new LarkNotifier.
// webhookURL must be a valid Lark bot webhook URL.
func NewLarkNotifier(webhookURL string) (*LarkNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("lark: webhook URL must not be empty")
	}
	return &LarkNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}, nil
}

type larkContent struct {
	Tag  string `json:"tag"`
	Text string `json:"text"`
}

type larkBody struct {
	MsgType string          `json:"msg_type"`
	Content larkTextContent `json:"content"`
}

type larkTextContent struct {
	Text string `json:"text"`
}

// Send delivers an alert to the configured Lark webhook.
func (n *LarkNotifier) Send(a Alert) error {
	text := fmt.Sprintf("[%s] VaultWatch Alert\nSecret: %s\nExpires: %s\nTime Left: %s",
		a.Level,
		a.SecretPath,
		a.ExpiresAt.Format("2006-01-02 15:04:05 UTC"),
		a.TimeLeft.String(),
	)

	payload := larkBody{
		MsgType: "text",
		Content: larkTextContent{Text: text},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("lark: failed to marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("lark: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("lark: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
