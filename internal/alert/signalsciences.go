package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// SignalSciencesNotifier sends alerts to Signal Sciences (Fastly Next-Gen WAF)
// via the custom alert API.
type SignalSciencesNotifier struct {
	corpName  string
	siteName  string
	apiToken  string
	endpoint  string
	client    *http.Client
}

type signalSciencesPayload struct {
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Source   string `json:"source"`
	Timestamp string `json:"timestamp"`
}

// NewSignalSciencesNotifier creates a new SignalSciencesNotifier.
// corpName and siteName identify the Signal Sciences corp/site.
// apiToken is the API access token.
func NewSignalSciencesNotifier(corpName, siteName, apiToken string) (*SignalSciencesNotifier, error) {
	if corpName == "" {
		return nil, fmt.Errorf("signal sciences: corp name must not be empty")
	}
	if siteName == "" {
		return nil, fmt.Errorf("signal sciences: site name must not be empty")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("signal sciences: api token must not be empty")
	}
	endpoint := fmt.Sprintf(
		"https://dashboard.signalsciences.net/api/v0/corps/%s/sites/%s/alerts",
		corpName, siteName,
	)
	return &SignalSciencesNotifier{
		corpName: corpName,
		siteName: siteName,
		apiToken: apiToken,
		endpoint: endpoint,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send delivers the alert to Signal Sciences.
func (n *SignalSciencesNotifier) Send(a Alert) error {
	severity := "info"
	if a.Level == LevelCritical {
		severity = "critical"
	} else if a.Level == LevelWarning {
		severity = "warning"
	}

	payload := signalSciencesPayload{
		Message:   a.String(),
		Severity:  severity,
		Source:    "vaultwatch",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("signal sciences: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, n.endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("signal sciences: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-user", "vaultwatch")
	req.Header.Set("x-api-token", n.apiToken)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("signal sciences: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("signal sciences: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
