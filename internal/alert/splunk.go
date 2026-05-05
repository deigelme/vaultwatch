package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// splunkEvent represents the Splunk HTTP Event Collector (HEC) payload.
type splunkEvent struct {
	Time   int64            `json:"time"`
	Source string           `json:"source"`
	Event  splunkEventBody  `json:"event"`
}

type splunkEventBody struct {
	Severity  string `json:"severity"`
	Message   string `json:"message"`
	Secret    string `json:"secret"`
	ExpiresIn string `json:"expires_in"`
}

// SplunkNotifier sends alerts to a Splunk HEC endpoint.
type SplunkNotifier struct {
	hecURL string
	token  string
	client *http.Client
}

// NewSplunkNotifier creates a new SplunkNotifier.
// hecURL should be the full HEC endpoint, e.g. https://splunk.example.com:8088/services/collector.
func NewSplunkNotifier(hecURL, token string) (*SplunkNotifier, error) {
	if hecURL == "" {
		return nil, fmt.Errorf("splunk: HEC URL must not be empty")
	}
	if token == "" {
		return nil, fmt.Errorf("splunk: HEC token must not be empty")
	}
	return &SplunkNotifier{
		hecURL: hecURL,
		token:  token,
		client: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers an Alert to Splunk via the HTTP Event Collector.
func (s *SplunkNotifier) Send(a Alert) error {
	payload := splunkEvent{
		Time:   time.Now().Unix(),
		Source: "vaultwatch",
		Event: splunkEventBody{
			Severity:  string(a.Level),
			Message:   a.Message,
			Secret:    a.SecretPath,
			ExpiresIn: a.TimeLeft.String(),
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("splunk: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, s.hecURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("splunk: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Splunk "+s.token)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("splunk: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("splunk: unexpected status %d", resp.StatusCode)
	}
	return nil
}
