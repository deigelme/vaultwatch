package alert

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// TelegramNotifier sends alerts via the Telegram Bot API.
type TelegramNotifier struct {
	botToken string
	chatID   string
	apiBase  string
	client   *http.Client
}

// NewTelegramNotifier creates a TelegramNotifier.
// botToken and chatID must both be non-empty.
func NewTelegramNotifier(botToken, chatID string) (*TelegramNotifier, error) {
	if botToken == "" {
		return nil, fmt.Errorf("telegram bot token must not be empty")
	}
	if chatID == "" {
		return nil, fmt.Errorf("telegram chat ID must not be empty")
	}
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		apiBase:  "https://api.telegram.org",
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type telegramResponse struct {
	OK bool `json:"ok"`
}

// Send posts a message to the configured Telegram chat.
func (t *TelegramNotifier) Send(a Alert) error {
	text := fmt.Sprintf("*[VaultWatch %s]*\n%s", a.Level, a.String())

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", t.apiBase, t.botToken)
	params := url.Values{}
	params.Set("chat_id", t.chatID)
	params.Set("text", text)
	params.Set("parse_mode", "Markdown")

	resp, err := t.client.PostForm(endpoint, params)
	if err != nil {
		return fmt.Errorf("telegram: send request: %w", err)
	}
	defer resp.Body.Close()

	var result telegramResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("telegram: decode response: %w", err)
	}
	if !result.OK {
		return fmt.Errorf("telegram: API returned ok=false (HTTP %d)", resp.StatusCode)
	}
	return nil
}
