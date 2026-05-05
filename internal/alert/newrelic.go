package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const newRelicEventsURL = "https://insights-collector.newrelic.com/v1/accounts/%s/events"

type newRelicEvent struct {
	EventType  string `json:"eventType"`
	Severity   string `json:"severity"`
	Message    string `json:"message"`
	SecretPath string `json:"secretPath"`
	ExpiresIn  string `json:"expiresIn"`
}

// NewRelicNotifier sends custom events to New Relic Insights.
type NewRelicNotifier struct {
	url       string
	apiKey    string
	client    *http.Client
}

// NewNewRelicNotifier creates a NewRelicNotifier.
// accountID is the New Relic account ID; apiKey is the Insights Insert API key.
func NewNewRelicNotifier(accountID, apiKey string) (*NewRelicNotifier, error) {
	if accountID == "" {
		return nil, fmt.Errorf("newrelic: account ID must not be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("newrelic: API key must not be empty")
	}
	return &NewRelicNotifier{
		url:    fmt.Sprintf(newRelicEventsURL, accountID),
		apiKey: apiKey,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers an Alert as a custom event to New Relic Insights.
func (n *NewRelicNotifier) Send(a Alert) error {
	evt := newRelicEvent{
		EventType:  "VaultWatchAlert",
		Severity:   string(a.Level),
		Message:    a.Message,
		SecretPath: a.SecretPath,
		ExpiresIn:  a.TimeLeft.String(),
	}

	body, err := json.Marshal([]newRelicEvent{evt})
	if err != nil {
		return fmt.Errorf("newrelic: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, n.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("newrelic: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Insert-Key", n.apiKey)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("newrelic: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("newrelic: unexpected status %d", resp.StatusCode)
	}
	return nil
}
