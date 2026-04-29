package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends alerts to PagerDuty via the Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	httpClient     *http.Client
	eventURL       string
}

type pdPayload struct {
	Summary   string `json:"summary"`
	Severity  string `json:"severity"`
	Timestamp string `json:"timestamp"`
	Source    string `json:"source"`
}

type pdEvent struct {
	RoutingKey  string    `json:"routing_key"`
	EventAction string    `json:"event_action"`
	Payload     pdPayload `json:"payload"`
}

// NewPagerDutyNotifier creates a PagerDutyNotifier with the given integration key.
func NewPagerDutyNotifier(integrationKey string) (*PagerDutyNotifier, error) {
	if integrationKey == "" {
		return nil, fmt.Errorf("pagerduty: integration key must not be empty")
	}
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		httpClient:     &http.Client{Timeout: 10 * time.Second},
		eventURL:       pagerDutyEventURL,
	}, nil
}

// Send dispatches an Alert to PagerDuty as a trigger event.
func (p *PagerDutyNotifier) Send(a Alert) error {
	severity := "warning"
	if a.Level == Critical {
		severity = "critical"
	}

	event := pdEvent{
		RoutingKey:  p.integrationKey,
		EventAction: "trigger",
		Payload: pdPayload{
			Summary:   a.String(),
			Severity:  severity,
			Timestamp: time.Now().UTC().Format(time.RFC3339),
			Source:    "vaultwatch",
		},
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("pagerduty: failed to marshal event: %w", err)
	}

	resp, err := p.httpClient.Post(p.eventURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("pagerduty: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("pagerduty: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
