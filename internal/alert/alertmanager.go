package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AlertmanagerNotifier sends alerts to a Prometheus Alertmanager instance.
type AlertmanagerNotifier struct {
	endpoint string
	client   *http.Client
}

type amAlert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	StartsAt    string            `json:"startsAt"`
}

// NewAlertmanagerNotifier creates an AlertmanagerNotifier.
// endpoint should be the full Alertmanager API URL, e.g.
// "http://alertmanager:9093/api/v2/alerts".
func NewAlertmanagerNotifier(endpoint string) (*AlertmanagerNotifier, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("alertmanager: endpoint must not be empty")
	}
	return &AlertmanagerNotifier{
		endpoint: endpoint,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send posts an alert to Alertmanager.
func (am *AlertmanagerNotifier) Send(a Alert) error {
	severity := "warning"
	if a.Level == Critical {
		severity = "critical"
	}

	payload := []amAlert{
		{
			Labels: map[string]string{
				"alertname": "VaultSecretExpiringSoon",
				"severity":  severity,
				"secret":    a.SecretPath,
				"source":    "vaultwatch",
			},
			Annotations: map[string]string{
				"summary":     a.String(),
				"time_left":   a.TimeLeft.String(),
			},
			StartsAt: time.Now().UTC().Format(time.RFC3339),
		},
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("alertmanager: marshal payload: %w", err)
	}

	resp, err := am.client.Post(am.endpoint, "application/json", bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("alertmanager: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("alertmanager: unexpected status %d", resp.StatusCode)
	}
	return nil
}
