package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const datadogEventsURL = "https://api.datadoghq.com/api/v1/events"

// DatadogNotifier sends alerts to Datadog as events.
type DatadogNotifier struct {
	apiKey  string
	apiURL  string
	client  *http.Client
}

type datadogEvent struct {
	Title      string   `json:"title"`
	Text       string   `json:"text"`
	AlertType  string   `json:"alert_type"`
	Tags       []string `json:"tags"`
	SourceType string   `json:"source_type_name"`
}

// NewDatadogNotifier creates a DatadogNotifier with the given API key.
// apiURL is optional; if empty the default Datadog events endpoint is used.
func NewDatadogNotifier(apiKey, apiURL string) (*DatadogNotifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("datadog: api key must not be empty")
	}
	if apiURL == "" {
		apiURL = datadogEventsURL
	}
	return &DatadogNotifier{
		apiKey: apiKey,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers an Alert to Datadog as an event.
func (d *DatadogNotifier) Send(a Alert) error {
	alertType := "info"
	switch a.Level {
	case LevelWarning:
		alertType = "warning"
	case LevelCritical:
		alertType = "error"
	}

	event := datadogEvent{
		Title:      fmt.Sprintf("VaultWatch: %s", a.SecretPath),
		Text:       a.String(),
		AlertType:  alertType,
		Tags:       []string{"source:vaultwatch", fmt.Sprintf("secret:%s", a.SecretPath)},
		SourceType: "vaultwatch",
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("datadog: failed to marshal event: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, d.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("datadog: failed to build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("DD-API-KEY", d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return fmt.Errorf("datadog: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("datadog: unexpected status %d", resp.StatusCode)
	}
	return nil
}
