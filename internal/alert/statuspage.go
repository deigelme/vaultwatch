package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// StatusPageNotifier sends incident updates to an Atlassian Statuspage component.
type StatusPageNotifier struct {
	pageID     string
	componentID string
	apiKey     string
	baseURL    string
	client     *http.Client
}

type statusPageBody struct {
	Component statusPageComponent `json:"component"`
}

type statusPageComponent struct {
	Status string `json:"status"`
}

// NewStatusPageNotifier returns a StatusPageNotifier or an error if required fields are missing.
func NewStatusPageNotifier(pageID, componentID, apiKey string) (*StatusPageNotifier, error) {
	if pageID == "" {
		return nil, fmt.Errorf("statuspage: page_id is required")
	}
	if componentID == "" {
		return nil, fmt.Errorf("statuspage: component_id is required")
	}
	if apiKey == "" {
		return nil, fmt.Errorf("statuspage: api_key is required")
	}
	return &StatusPageNotifier{
		pageID:      pageID,
		componentID: componentID,
		apiKey:      apiKey,
		baseURL:     "https://api.statuspage.io/v1",
		client:      &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send updates the Statuspage component status based on the alert level.
func (n *StatusPageNotifier) Send(a Alert) error {
	status := "operational"
	switch a.Level {
	case LevelWarning:
		status = "degraded_performance"
	case LevelCritical:
		status = "major_outage"
	}

	body, err := json.Marshal(statusPageBody{Component: statusPageComponent{Status: status}})
	if err != nil {
		return fmt.Errorf("statuspage: marshal: %w", err)
	}

	url := fmt.Sprintf("%s/pages/%s/components/%s", n.baseURL, n.pageID, n.componentID)
	req, err := http.NewRequest(http.MethodPatch, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("statuspage: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "OAuth "+n.apiKey)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("statuspage: send: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("statuspage: unexpected status %d", resp.StatusCode)
	}
	return nil
}
