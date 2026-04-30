package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// victorOpsPayload represents the VictorOps (Splunk On-Call) alert payload.
type victorOpsPayload struct {
	MessageType       string `json:"message_type"`
	EntityID          string `json:"entity_id"`
	EntityDisplayName string `json:"entity_display_name"`
	StateMessage      string `json:"state_message"`
	Timestamp         int64  `json:"timestamp"`
}

// VictorOpsNotifier sends alerts to VictorOps (Splunk On-Call) via REST endpoint.
type VictorOpsNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewVictorOpsNotifier creates a new VictorOpsNotifier.
// webhookURL is the VictorOps REST endpoint URL including the routing key.
func NewVictorOpsNotifier(webhookURL string) (*VictorOpsNotifier, error) {
	if webhookURL == "" {
		return nil, fmt.Errorf("victorops: webhook URL must not be empty")
	}
	return &VictorOpsNotifier{
		webhookURL: webhookURL,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send dispatches an alert to VictorOps.
func (v *VictorOpsNotifier) Send(a Alert) error {
	msgType := "WARNING"
	if a.Level == Critical {
		msgType = "CRITICAL"
	}

	payload := victorOpsPayload{
		MessageType:       msgType,
		EntityID:          a.SecretPath,
		EntityDisplayName: fmt.Sprintf("VaultWatch: %s", a.SecretPath),
		StateMessage:      a.String(),
		Timestamp:         time.Now().Unix(),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("victorops: failed to marshal payload: %w", err)
	}

	resp, err := v.client.Post(v.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("victorops: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("victorops: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
