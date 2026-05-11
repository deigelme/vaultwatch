package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// freshServiceTicket is the payload sent to the Freshservice Tickets API.
type freshServiceTicket struct {
	Subject     string `json:"subject"`
	Description string `json:"description"`
	Email       string `json:"email"`
	Priority    int    `json:"priority"`
	Status      int    `json:"status"`
}

// freshServiceNotifier sends alerts to Freshservice by creating support tickets.
type freshServiceNotifier struct {
	baseURL  string
	apiKey   string
	reporter string
	client   *http.Client
}

// freshServicePriority maps an alert Level to a Freshservice ticket priority.
// 1=Low, 2=Medium, 3=High, 4=Urgent.
func freshServicePriority(level Level) int {
	switch level {
	case LevelCritical:
		return 4 // Urgent
	case LevelWarning:
		return 3 // High
	default:
		return 2 // Medium
	}
}

// NewFreshserviceNotifier constructs a notifier that creates Freshservice tickets.
// baseURL is the Freshservice domain root (e.g. "https://yourcompany.freshservice.com").
// apiKey is the Freshservice API key used for HTTP Basic auth.
// reporter is the email address shown as the ticket requester.
func NewFreshserviceNotifier(baseURL, apiKey, reporter string) (*freshServiceNotifier, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("freshservice: baseURL must not be empty")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("freshservice: apiKey must not be empty")
	}
	if reporter == "" {
		return nil, fmt.Errorf("freshservice: reporter email must not be empty")
	}
	return &freshServiceNotifier{
		baseURL:  baseURL,
		apiKey:   apiKey,
		reporter: reporter,
		client:   &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send creates a Freshservice ticket for the given alert.
func (n *freshServiceNotifier) Send(a Alert) error {
	ticket := freshServiceTicket{
		Subject:     fmt.Sprintf("[VaultWatch] %s", a.SecretPath),
		Description: a.String(),
		Email:       n.reporter,
		Priority:    freshServicePriority(a.Level),
		Status:      2, // Open
	}
	body, err := json.Marshal(map[string]any{"ticket": ticket})
	if err != nil {
		return fmt.Errorf("freshservice: marshal payload: %w", err)
	}
	url := n.baseURL + "/api/v2/tickets"
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("freshservice: build request: %w", err)
	}
	req.SetBasicAuth(n.apiKey, "X")
	req.Header.Set("Content-Type", "application/json")
	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("freshservice: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("freshservice: unexpected status %d", resp.StatusCode)
	}
	return nil
}
