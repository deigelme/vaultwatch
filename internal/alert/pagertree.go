package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// PagerTreeNotifier sends alerts to PagerTree via its integration API.
type PagerTreeNotifier struct {
	integrationURL string
	client         *http.Client
}

// NewPagerTreeNotifier creates a new PagerTreeNotifier.
// integrationURL is the full PagerTree integration endpoint URL.
func NewPagerTreeNotifier(integrationURL string) (*PagerTreeNotifier, error) {
	if integrationURL == "" {
		return nil, fmt.Errorf("pagertree: integration URL must not be empty")
	}
	return &PagerTreeNotifier{
		integrationURL: integrationURL,
		client:         &http.Client{},
	}, nil
}

type pagerTreePayload struct {
	Title   string `json:"title"`
	Details string `json:"details"`
	Urgency string `json:"urgency"`
}

func pagerTreeUrgency(level Level) string {
	switch level {
	case LevelCritical:
		return "critical"
	case LevelWarning:
		return "medium"
	default:
		return "low"
	}
}

// Send delivers an alert to PagerTree.
func (n *PagerTreeNotifier) Send(a Alert) error {
	payload := pagerTreePayload{
		Title:   fmt.Sprintf("[%s] %s", a.Level, a.SecretPath),
		Details: a.String(),
		Urgency: pagerTreeUrgency(a.Level),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("pagertree: failed to marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.integrationURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("pagertree: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("pagertree: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
