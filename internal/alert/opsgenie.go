package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// OpsGenieNotifier sends alerts to OpsGenie.
type OpsGenieNotifier struct {
	apiKey   string
	baseURL  string
	client   *http.Client
}

// NewOpsGenieNotifier creates a new OpsGenieNotifier.
func NewOpsGenieNotifier(apiKey, baseURL string) (*OpsGenieNotifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("opsgenie: api key must not be empty")
	}
	if baseURL == "" {
		baseURL = "https://api.opsgenie.com"
	}
	return &OpsGenieNotifier{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  &http.Client{},
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

// Send delivers the alert to OpsGenie.
func (n *OpsGenieNotifier) Send(a Alert) error {
	payload := map[string]interface{}{
		"message":     a.String(),
		"description": fmt.Sprintf("Secret: %s | Expires: %s", a.SecretPath, a.ExpiresAt.Format("2006-01-02 15:04:05")),
		"priority":    opsGeniePriority(a.Level),
		"tags":        []string{"vaultwatch"},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("opsgenie: marshal payload: %w", err)
	}
	req, err := http.NewRequest(http.MethodPost, n.baseURL+"/v2/alerts", bytes.NewReader(body))
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
