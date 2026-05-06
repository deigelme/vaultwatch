package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const defaultOpsGenieEndpoint = "https://api.opsgenie.com/v2/alerts"

// OpsGenieNotifier sends alerts to OpsGenie.
type OpsGenieNotifier struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// NewOpsGenieNotifier creates a new OpsGenieNotifier.
// apiKey is required; endpoint is optional (defaults to OpsGenie cloud API).
func NewOpsGenieNotifier(apiKey, endpoint string) (*OpsGenieNotifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("opsgenie: api key must not be empty")
	}
	if endpoint == "" {
		endpoint = defaultOpsGenieEndpoint
	}
	return &OpsGenieNotifier{
		apiKey:   apiKey,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

type opsGeniePayload struct {
	Message     string            `json:"message"`
	Description string            `json:"description"`
	Priority    string            `json:"priority"`
	Details     map[string]string `json:"details,omitempty"`
}

func opsGeniePriority(level Level) string {
	switch level {
	case LevelCritical:
		return "P1"
	case LevelWarning:
		return "P3"
	default:
		return "P5"
	}
}

// Send delivers the alert to OpsGenie.
func (n *OpsGenieNotifier) Send(a Alert) error {
	payload := opsGeniePayload{
		Message:     fmt.Sprintf("[%s] Vault secret expiring: %s", a.Level, a.SecretPath),
		Description: a.String(),
		Priority:    opsGeniePriority(a.Level),
		Details: map[string]string{
			"secret_path": a.SecretPath,
			"expires_in":  a.TimeLeft.String(),
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, n.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "GenieKey "+n.apiKey)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("opsgenie: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("opsgenie: unexpected status %d", resp.StatusCode)
	}
	return nil
}
