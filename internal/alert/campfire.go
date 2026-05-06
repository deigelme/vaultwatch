package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// CampfireNotifier sends alerts to a Basecamp Campfire room via the
// Basecamp 3 Chatbot API.
type CampfireNotifier struct {
	accountID string
	campfireID string
	token string
	httpClient *http.Client
	baseURL string // overridable for testing
}

type campfirePayload struct {
	Content string `json:"content"`
}

// NewCampfireNotifier creates a CampfireNotifier.
// accountID is the Basecamp account ID, campfireID is the Campfire room ID,
// and token is a Basecamp API access token.
func NewCampfireNotifier(accountID, campfireID, token string) (*CampfireNotifier, error) {
	if strings.TrimSpace(accountID) == "" {
		return nil, fmt.Errorf("campfire: accountID must not be empty")
	}
	if strings.TrimSpace(campfireID) == "" {
		return nil, fmt.Errorf("campfire: campfireID must not be empty")
	}
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("campfire: token must not be empty")
	}
	return &CampfireNotifier{
		accountID:  accountID,
		campfireID: campfireID,
		token:      token,
		httpClient: &http.Client{},
		baseURL:    "https://3.basecampapi.com",
	}, nil
}

// Send posts an alert message to the configured Campfire room.
func (c *CampfireNotifier) Send(a Alert) error {
	url := fmt.Sprintf("%s/%s/integrations/%s/buckets/%s/chats/%s/lines.json",
		c.baseURL, c.accountID, c.token, c.accountID, c.campfireID)

	body, err := json.Marshal(campfirePayload{
		Content: fmt.Sprintf("[%s] %s", a.Level, a.Message),
	})
	if err != nil {
		return fmt.Errorf("campfire: marshal payload: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("campfire: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "vaultwatch/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("campfire: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("campfire: unexpected status %d", resp.StatusCode)
	}
	return nil
}
