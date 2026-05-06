package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// serviceNowUrgency maps alert levels to ServiceNow urgency values (1=High, 2=Medium, 3=Low).
func serviceNowUrgency(level Level) string {
	switch level {
	case Critical:
		return "1"
	case Warning:
		return "2"
	default:
		return "3"
	}
}

type serviceNowNotifier struct {
	baseURL  string
	username string
	password string
	table    string
	client   *http.Client
}

// NewServiceNowNotifier creates a Notifier that opens incidents in ServiceNow.
// baseURL should be the root of the ServiceNow instance, e.g. https://dev12345.service-now.com.
func NewServiceNowNotifier(baseURL, username, password, table string) (Notifier, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("servicenow: base URL must not be empty")
	}
	if strings.TrimSpace(username) == "" || strings.TrimSpace(password) == "" {
		return nil, fmt.Errorf("servicenow: username and password must not be empty")
	}
	if strings.TrimSpace(table) == "" {
		table = "incident"
	}
	return &serviceNowNotifier{
		baseURL:  strings.TrimRight(baseURL, "/"),
		username: username,
		password: password,
		table:    table,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

func (s *serviceNowNotifier) Send(a Alert) error {
	payload := map[string]string{
		"short_description": fmt.Sprintf("[VaultWatch] %s", a.Message),
		"description":       fmt.Sprintf("Secret path: %s\nExpires at: %s\nTime left: %s", a.Path, a.ExpiresAt.Format(time.RFC3339), a.TimeLeft.Round(time.Second)),
		"urgency":           serviceNowUrgency(a.Level),
		"category":          "Security",
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("servicenow: failed to marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/api/now/table/%s", s.baseURL, s.table)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("servicenow: failed to build request: %w", err)
	}
	req.SetBasicAuth(s.username, s.password)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("servicenow: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusOK {
		return fmt.Errorf("servicenow: unexpected status %d", resp.StatusCode)
	}
	return nil
}
