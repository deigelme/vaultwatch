package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GotifyNotifier sends alerts to a self-hosted Gotify server.
type GotifyNotifier struct {
	baseURL  string
	token    string
	client   *http.Client
}

type gotifyPayload struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Priority int    `json:"priority"`
}

// NewGotifyNotifier creates a GotifyNotifier.
// baseURL is the root URL of the Gotify server (e.g. https://gotify.example.com).
// token is the application token used to publish messages.
func NewGotifyNotifier(baseURL, token string) (*GotifyNotifier, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("gotify: baseURL must not be empty")
	}
	if token == "" {
		return nil, fmt.Errorf("gotify: token must not be empty")
	}
	return &GotifyNotifier{
		baseURL: baseURL,
		token:   token,
		client:  &http.Client{},
	}, nil
}

// Send publishes an alert message to Gotify.
func (g *GotifyNotifier) Send(a Alert) error {
	priority := 5
	if a.Level == LevelCritical {
		priority = 10
	}

	payload := gotifyPayload{
		Title:    fmt.Sprintf("VaultWatch [%s]", a.Level),
		Message:  a.String(),
		Priority: priority,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("gotify: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/message?token=%s", g.baseURL, g.token)
	resp, err := g.client.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("gotify: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("gotify: unexpected status %d", resp.StatusCode)
	}
	return nil
}
