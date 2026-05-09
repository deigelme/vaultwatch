package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const amplitudeDefaultEndpoint = "https://api2.amplitude.com/2/httpapi"

// AmplitudeNotifier sends VaultWatch alerts as events to Amplitude Analytics.
type AmplitudeNotifier struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// NewAmplitudeNotifier creates an AmplitudeNotifier.
// apiKey is required. endpoint is optional; the Amplitude HTTP API v2 URL is
// used when empty.
func NewAmplitudeNotifier(apiKey, endpoint string) (*AmplitudeNotifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("amplitude: api key must not be empty")
	}
	ep := endpoint
	if ep == "" {
		ep = amplitudeDefaultEndpoint
	}
	return &AmplitudeNotifier{
		apiKey:   apiKey,
		endpoint: ep,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers an alert event to Amplitude.
func (n *AmplitudeNotifier) Send(a Alert) error {
	type eventProps struct {
		SecretPath string `json:"secret_path"`
		Level      string `json:"level"`
		Message    string `json:"message"`
		TimeLeft   string `json:"time_left"`
	}
	type event struct {
		UserID          string     `json:"user_id"`
		EventType       string     `json:"event_type"`
		EventProperties eventProps `json:"event_properties"`
	}
	type payload struct {
		APIKey string  `json:"api_key"`
		Events []event `json:"events"`
	}

	p := payload{
		APIKey: n.apiKey,
		Events: []event{
			{
				UserID:    "vaultwatch",
				EventType: "secret_expiry_alert",
				EventProperties: eventProps{
					SecretPath: a.SecretPath,
					Level:      a.Level.String(),
					Message:    a.Message,
					TimeLeft:   a.TimeLeft.String(),
				},
			},
		},
	}

	body, err := json.Marshal(p)
	if err != nil {
		return fmt.Errorf("amplitude: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.endpoint, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("amplitude: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("amplitude: unexpected status %d", resp.StatusCode)
	}
	return nil
}
