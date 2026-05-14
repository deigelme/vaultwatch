package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OpsGenieNotifier sends alerts to OpsGenie via its REST API.
type OpsGenieNotifier struct {
	apiKey   string
	endpoint string
	client   *http.Client
}

// NewOpsGenieNotifier creates a new OpsGenieNotifier.
// apiKey must be non-empty. endpoint defaults to the OpsGenie API if empty.
func NewOpsGenieNotifier(apiKey, endpoint string) (*OpsGenieNotifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("opsgenie: api key must not be empty")
	}
	if endpoint == "" {
		endpoint = "https://api.opsgenie.com/v2/alerts"
	}
	return &OpsGenieNotifier{
		apiKey:   apiKey,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
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

// Send dispatches an alert to OpsGenie.
func (n *OpsGenieNotifier) Send(a Alert) error {
	payload := map[string]interface{}{
		"message":     a.String(),
		"description": fmt.Sprintf("Secret %s expires at %s", a.SecretPath, a.ExpiresAt.Format(time.RFC3339)),
		"priority":    opsGeniePriority(a.Level),
		"tags":        []string{"vaultwatch"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, n.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("opsgenie: build request: %w", err)
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
