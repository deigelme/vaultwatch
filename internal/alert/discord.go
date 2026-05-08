package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// DiscordNotifier sends alerts to a Discord channel via webhook.
type DiscordNotifier struct {
	webhookURL string
	client     *http.Client
}

type discordPayload struct {
	Username string         `json:"username"`
	Embeds   []discordEmbed `json:"embeds"`
}

type discordEmbed struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Color       int    `json:"color"`
}

// NewDiscordNotifier creates a DiscordNotifier. Returns an error if webhookURL is empty.
func NewDiscordNotifier(webhookURL string) (*DiscordNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("discord webhook URL must not be empty")
	}
	return &DiscordNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send posts an alert embed to the configured Discord webhook.
func (d *DiscordNotifier) Send(a Alert) error {
	color := 0x00bfff // info blue
	if a.Level == LevelCritical {
		color = 0xff0000 // red
	} else if a.Level == LevelWarning {
		color = 0xffa500 // orange
	}

	payload := discordPayload{
		Username: "VaultWatch",
		Embeds: []discordEmbed{
			{
				Title:       fmt.Sprintf("[%s] Vault Secret Expiring", a.Level),
				Description: a.String(),
				Color:       color,
			},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("discord: marshal payload: %w", err)
	}

	resp, err := d.client.Post(d.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("discord: unexpected status %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
	}
	return nil
}
