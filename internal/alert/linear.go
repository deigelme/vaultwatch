package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// linearIssuePayload represents the GraphQL mutation payload for creating a Linear issue.
type linearIssuePayload struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables"`
}

// LinearNotifier sends alerts to Linear by creating issues via the Linear GraphQL API.
type LinearNotifier struct {
	apiKey    string
	teamID    string
	labelIDs  []string
	assigneeID string
	client    *http.Client
}

// NewLinearNotifier creates a new LinearNotifier.
// apiKey is the Linear personal API key, teamID is the team to create issues in.
// labelIDs and assigneeID are optional.
func NewLinearNotifier(apiKey, teamID string, labelIDs []string, assigneeID string) (*LinearNotifier, error) {
	if apiKey == "" {
		return nil, fmt.Errorf("linear: API key must not be empty")
	}
	if teamID == "" {
		return nil, fmt.Errorf("linear: team ID must not be empty")
	}
	return &LinearNotifier{
		apiKey:     apiKey,
		teamID:     teamID,
		labelIDs:   labelIDs,
		assigneeID: assigneeID,
		client:     &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// linearPriority maps an alert Level to a Linear issue priority (0=no, 1=urgent, 2=high, 3=medium, 4=low).
func linearPriority(level Level) int {
	switch level {
	case LevelCritical:
		return 1 // Urgent
	case LevelWarning:
		return 2 // High
	default:
		return 3 // Medium
	}
}

// Send creates a Linear issue for the given alert.
func (n *LinearNotifier) Send(a Alert) error {
	mutation := `
		mutation IssueCreate($input: IssueCreateInput!) {
			issueCreate(input: $input) {
				success
				issue {
					id
					title
				}
			}
		}`

	input := map[string]interface{}{
		"title":       fmt.Sprintf("[VaultWatch] %s", a.SecretPath),
		"description": fmt.Sprintf("%s\n\nSecret: `%s`\nExpires: %s\nTime left: %s",
			a.Message,
			a.SecretPath,
			a.ExpiresAt.UTC().Format(time.RFC3339),
			a.TimeLeft.Round(time.Minute).String(),
		),
		"teamId":   n.teamID,
		"priority": linearPriority(a.Level),
	}

	if len(n.labelIDs) > 0 {
		input["labelIds"] = n.labelIDs
	}
	if n.assigneeID != "" {
		input["assigneeId"] = n.assigneeID
	}

	payload := linearIssuePayload{
		Query:     mutation,
		Variables: map[string]interface{}{"input": input},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("linear: failed to marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, "https://api.linear.app/graphql", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("linear: failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", n.apiKey)

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("linear: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("linear: unexpected status code %d", resp.StatusCode)
	}

	var result struct {
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Errorf("linear: failed to decode response: %w", err)
	}
	if len(result.Errors) > 0 {
		return fmt.Errorf("linear: API error: %s", result.Errors[0].Message)
	}

	return nil
}
