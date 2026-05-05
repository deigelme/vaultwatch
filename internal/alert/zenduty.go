package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ZendutyNotifier sends alerts to Zenduty via its Events API.
type ZendutyNotifier struct {
	integrationKey string
	apiURL         string
	client         *http.Client
}

type zendutyPayload struct {
	AlertType   string            `json:"alert_type"`
	Message     string            `json:"message"`
	Summary     string            `json:"summary"`
	EntityID    string            `json:"entity_id"`
	Payload     map[string]string `json:"payload"`
	CreatedAt   string            `json:"created_at"`
}

// NewZendutyNotifier creates a ZendutyNotifier.
// integrationKey is the Zenduty service integration key.
func NewZendutyNotifier(integrationKey string) (*ZendutyNotifier, error) {
	if integrationKey == "" {
		return nil, fmt.Errorf("zenduty: integration key must not be empty")
	}
	return &ZendutyNotifier{
		integrationKey: integrationKey,
		apiURL:         "https://www.zenduty.com/api/events/" + integrationKey + "/",
		client:         &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send dispatches an Alert to Zenduty.
func (z *ZendutyNotifier) Send(a Alert) error {
	alertType := "info"
	if a.Level == Critical {
		alertType = "critical"
	} else if a.Level == Warning {
		alertType = "warning"
	}

	body := zendutyPayload{
		AlertType: alertType,
		Message:   a.String(),
		Summary:   fmt.Sprintf("VaultWatch: secret %q expires soon", a.SecretPath),
		EntityID:  a.SecretPath,
		Payload: map[string]string{
			"secret_path": a.SecretPath,
			"time_left":   a.TimeLeft.String(),
			"level":       a.Level.String(),
		},
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("zenduty: marshal payload: %w", err)
	}

	resp, err := z.client.Post(z.apiURL, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("zenduty: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("zenduty: unexpected status %d", resp.StatusCode)
	}
	return nil
}
