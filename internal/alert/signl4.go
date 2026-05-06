package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SIGNL4Notifier sends alerts to SIGNL4 via its webhook API.
type SIGNL4Notifier struct {
	webhookURL string
	client     *http.Client
}

type signl4Payload struct {
	Title    string `json:"title"`
	Message  string `json:"message"`
	Severity int    `json:"severity"`
	Source   string `json:"source"`
}

// NewSIGNL4Notifier creates a new SIGNL4Notifier.
// webhookURL is the SIGNL4 team webhook URL (https://connect.signl4.com/webhook/<teamSecret>).
func NewSIGNL4Notifier(webhookURL string) (*SIGNL4Notifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("signl4: webhook URL must not be empty")
	}
	return &SIGNL4Notifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers an alert to SIGNL4.
func (n *SIGNL4Notifier) Send(a Alert) error {
	severity := 1 // low
	if a.Level == LevelCritical {
		severity = 3 // high
	} else if a.Level == LevelWarning {
		severity = 2 // medium
	}

	payload := signl4Payload{
		Title:    fmt.Sprintf("[%s] Vault Secret Expiring: %s", a.Level, a.SecretPath),
		Message:  a.String(),
		Severity: severity,
		Source:   "vaultwatch",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("signl4: failed to marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("signl4: HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("signl4: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
