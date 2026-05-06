package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ServiceNowNotifier sends alerts to a ServiceNow instance by creating
// incidents via the Table API.
type ServiceNowNotifier struct {
	baseURL    string
	username   string
	password   string
	assignedTo string
	client     *http.Client
}

// serviceNowIncident represents the payload sent to the ServiceNow Table API.
type serviceNowIncident struct {
	ShortDescription string `json:"short_description"`
	Description      string `json:"description"`
	Urgency          string `json:"urgency"`
	Impact           string `json:"impact"`
	AssignedTo       string `json:"assigned_to,omitempty"`
	Category         string `json:"category"`
}

// serviceNowUrgency maps alert levels to ServiceNow urgency values.
// 1 = High, 2 = Medium, 3 = Low.
func serviceNowUrgency(level Level) string {
	switch level {
	case LevelCritical:
		return "1"
	case LevelWarning:
		return "2"
	default:
		return "3"
	}
}

// NewServiceNowNotifier creates a ServiceNowNotifier that targets the given
// instance base URL (e.g. "https://dev12345.service-now.com") using HTTP Basic
// authentication. assignedTo may be empty if no specific assignment is needed.
func NewServiceNowNotifier(baseURL, username, password, assignedTo string) (*ServiceNowNotifier, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("servicenow: base URL must not be empty")
	}
	if username == "" || password == "" {
		return nil, fmt.Errorf("servicenow: username and password must not be empty")
	}
	return &ServiceNowNotifier{
		baseURL:    baseURL,
		username:   username,
		password:   password,
		assignedTo: assignedTo,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send creates a ServiceNow incident for the given alert.
func (n *ServiceNowNotifier) Send(a Alert) error {
	urgency := serviceNowUrgency(a.Level)

	incident := serviceNowIncident{
		ShortDescription: fmt.Sprintf("[VaultWatch] %s", a.SecretPath),
		Description:      a.String(),
		Urgency:          urgency,
		Impact:           urgency, // mirror urgency as impact
		AssignedTo:       n.assignedTo,
		Category:         "security",
	}

	body, err := json.Marshal(incident)
	if err != nil {
		return fmt.Errorf("servicenow: failed to marshal incident: %w", err)
	}

	url := fmt.Sprintf("%s/api/now/table/incident", n.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("servicenow: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(n.username, n.password)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("servicenow: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("servicenow: unexpected status code %d", resp.StatusCode)
	}
	return nil
}
