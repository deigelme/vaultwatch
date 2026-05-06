package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const pagerDutyEventsURL = "https://events.pagerduty.com/v2/enqueue"

// PagerDutyNotifier sends alerts to PagerDuty via the Events API v2.
type PagerDutyNotifier struct {
	integrationKey string
	endpoint       string
	client         *http.Client
}

// NewPagerDutyNotifier creates a PagerDutyNotifier.
// integrationKey is the PagerDuty Events API v2 integration key.
func NewPagerDutyNotifier(integrationKey string) (*PagerDutyNotifier, error) {
	if integrationKey == "" {
		return nil, fmt.Errorf("pagerduty: integration key must not be empty")
	}
	return &PagerDutyNotifier{
		integrationKey: integrationKey,
		endpoint:       pagerDutyEventsURL,
		client:         &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type pagerDutyPayload struct {
	RoutingKey  string            `json:"routing_key"`
	EventAction string            `json:"event_action"`
	Payload     pagerDutyInner    `json:"payload"`
}

type pagerDutyInner struct {
	Summary   string `json:"summary"`
	Severity  string `json:"severity"`
	Source    string `json:"source"`
	Timestamp string `json:"timestamp"`
}

func pagerDutySeverity(level Level) string {
	switch level {
	case LevelCritical:
		return "critical"
	case LevelWarning:
		return "warning"
	default:
		return "info"
	}
}

// Send dispatches an alert to PagerDuty.
func (n *PagerDutyNotifier) Send(a Alert) error {
	body := pagerDutyPayload{
		RoutingKey:  n.integrationKey,
		EventAction: "trigger",
		Payload: pagerDutyInner{
			Summary:   a.String(),
			Severity:  pagerDutySeverity(a.Level),
			Source:    "vaultwatch",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("pagerduty: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("pagerduty: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagerduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
