package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// JiraNotifier creates Jira issues for expiring secrets.
type JiraNotifier struct {
	baseURL   string
	username  string
	apiToken  string
	projectKey string
	issueType  string
	client    *http.Client
}

type jiraIssuePayload struct {
	Fields jiraFields `json:"fields"`
}

type jiraFields struct {
	Project   jiraProject   `json:"project"`
	Summary   string        `json:"summary"`
	Description string      `json:"description"`
	IssueType jiraIssueType `json:"issuetype"`
	Priority  jiraPriority  `json:"priority"`
}

type jiraProject struct {
	Key string `json:"key"`
}

type jiraIssueType struct {
	Name string `json:"name"`
}

type jiraPriority struct {
	Name string `json:"name"`
}

// NewJiraNotifier creates a new JiraNotifier.
func NewJiraNotifier(baseURL, username, apiToken, projectKey, issueType string) (*JiraNotifier, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("jira: baseURL must not be empty")
	}
	if username == "" || apiToken == "" {
		return nil, fmt.Errorf("jira: username and apiToken must not be empty")
	}
	if projectKey == "" {
		return nil, fmt.Errorf("jira: projectKey must not be empty")
	}
	it := issueType
	if it == "" {
		it = "Task"
	}
	return &JiraNotifier{
		baseURL:    baseURL,
		username:   username,
		apiToken:   apiToken,
		projectKey: projectKey,
		issueType:  it,
		client:     &http.Client{},
	}, nil
}

// Send creates a Jira issue for the given alert.
func (j *JiraNotifier) Send(a Alert) error {
	priority := jiraPriorityName(a.Level)
	payload := jiraIssuePayload{
		Fields: jiraFields{
			Project:     jiraProject{Key: j.projectKey},
			Summary:     fmt.Sprintf("[VaultWatch] %s", a.String()),
			Description: fmt.Sprintf("Secret path: %s\nExpires in: %s\nLevel: %s", a.SecretPath, a.TimeLeft, a.Level),
			IssueType:   jiraIssueType{Name: j.issueType},
			Priority:    jiraPriority{Name: priority},
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("jira: marshal payload: %w", err)
	}
	url := fmt.Sprintf("%s/rest/api/2/issue", j.baseURL)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("jira: create request: %w", err)
	}
	req.SetBasicAuth(j.username, j.apiToken)
	req.Header.Set("Content-Type", "application/json")
	resp, err := j.client.Do(req)
	if err != nil {
		return fmt.Errorf("jira: send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("jira: unexpected status %d", resp.StatusCode)
	}
	return nil
}

func jiraPriorityName(level Level) string {
	switch level {
	case LevelCritical:
		return "Highest"
	case LevelWarning:
		return "Medium"
	default:
		return "Low"
	}
}
