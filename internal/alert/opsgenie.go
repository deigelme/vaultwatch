package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultOpsGenieURL = "https://api.opsgenie.com/v2/alerts"

// OpsGenieNotifier sends alerts to OpsGenie.
type OpsGenieNotifier struct {
	apiKey  string
	apiURL  string
	client  *http.Client
}

type opsGeniePayload struct {
	Message     string            `json:"message"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Tags        []string          `json:"tags"`
	Details     map[string]string `json:"details"`
}

// NewOpsGenieNotifier creates a new OpsGenieNotifier.
// apiKey must be non-empty. apiURL is optional; defaults to the OpsGenie v2 alerts endpoint.
func NewOpsGenieNotifier(apiKey, apiURL string) (*OpsGenieNotifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("opsgenie: api key must not be empty")
	}
	if apiURL == "" {
		apiURL = defaultOpsGenieURL
	}
	return &OpsGenieNotifier{
		apiKey: apiKey,
		apiURL: apiURL,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers an Alert to OpsGenie.
func (o *OpsGenieNotifier) Send(a Alert) error {
	priority := opsGeniePriority(a.Level)

	payload := opsGeniePayload{
		Message:     fmt.Sprintf("VaultWatch: %s", a.SecretPath),
		Description: a.String(),
		Priority:    priority,
		Tags:        []string{"vaultwatch", string(a.Level)},
		Details: map[string]string{
			"secret_path": a.SecretPath,
			"expires_at":  a.ExpiresAt.Format(time.RFC3339),
			"time_left":   a.TimeLeft.String(),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, o.apiURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+o.apiKey)

	resp, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func opsGeniePriority(level Level) string {
	switch level {
	case LevelCritical:
		return "P1"
	case LevelWarning:
		return "P2"
	default:
		return "P3"
	}
}
