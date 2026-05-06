package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// GrafanaNotifier sends alerts to Grafana OnCall via its webhook integration.
type GrafanaNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewGrafanaNotifier creates a new GrafanaNotifier.
// webhookURL is the Grafana OnCall integration webhook URL.
func NewGrafanaNotifier(webhookURL string) (*GrafanaNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("grafana: webhook URL must not be empty")
	}
	return &GrafanaNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{},
	}, nil
}

type grafanaPayload struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	State    string `json:"state"`
	ImageURL string `json:"imageUrl,omitempty"`
}

func grafanaState(level Level) string {
	switch level {
	case LevelCritical:
		return "alerting"
	case LevelWarning:
		return "pending"
	default:
		return "ok"
	}
}

// Send delivers the alert to Grafana OnCall.
func (g *GrafanaNotifier) Send(a Alert) error {
	payload := grafanaPayload{
		Title:   fmt.Sprintf("[%s] VaultWatch: %s", a.Level, a.SecretPath),
		Message: a.String(),
		State:   grafanaState(a.Level),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("grafana: failed to marshal payload: %w", err)
	}

	resp, err := g.client.Post(g.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("grafana: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("grafana: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
