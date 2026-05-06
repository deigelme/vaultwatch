package alert

import (
	"bytes"
	"fmt"
	"net/http"
)

// NtfyNotifier sends alerts to an ntfy.sh topic (self-hosted or public).
type NtfyNotifier struct {
	baseURL string
	topic   string
	client  *http.Client
}

// NewNtfyNotifier creates a new NtfyNotifier.
// baseURL should be the server root, e.g. "https://ntfy.sh" or a self-hosted URL.
// topic is the ntfy topic name to publish to.
func NewNtfyNotifier(baseURL, topic string) (*NtfyNotifier, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("ntfy: baseURL must not be empty")
	}
	if topic == "" {
		return nil, fmt.Errorf("ntfy: topic must not be empty")
	}
	return &NtfyNotifier{
		baseURL: baseURL,
		topic:   topic,
		client:  &http.Client{},
	}, nil
}

// Send publishes the alert to the configured ntfy topic.
func (n *NtfyNotifier) Send(a Alert) error {
	url := fmt.Sprintf("%s/%s", n.baseURL, n.topic)

	body := []byte(a.String())

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("ntfy: failed to build request: %w", err)
	}

	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Title", fmt.Sprintf("VaultWatch — %s expiring", a.SecretPath))
	req.Header.Set("Priority", ntfyPriority(a.Level))
	req.Header.Set("Tags", "key,rotating_light")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("ntfy: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ntfy: unexpected status %d", resp.StatusCode)
	}
	return nil
}

// ntfyPriority maps an alert Level to an ntfy message priority string.
func ntfyPriority(level Level) string {
	switch level {
	case LevelCritical:
		return "urgent"
	case LevelWarning:
		return "high"
	default:
		return "default"
	}
}
