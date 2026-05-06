package alert

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

const pushoverAPIURL = "https://api.pushover.net/1/messages.json"

// PushoverNotifier sends alerts via the Pushover API.
type PushoverNotifier struct {
	appToken string
	userKey  string
	apiURL   string
	client   *http.Client
}

// NewPushoverNotifier creates a new PushoverNotifier.
// appToken is the Pushover application token and userKey is the recipient user/group key.
func NewPushoverNotifier(appToken, userKey string) (*PushoverNotifier, error) {
	if appToken == "" {
		return nil, fmt.Errorf("pushover: app token must not be empty")
	}
	if userKey == "" {
		return nil, fmt.Errorf("pushover: user key must not be empty")
	}
	return &PushoverNotifier{
		appToken: appToken,
		userKey:  userKey,
		apiURL:   pushoverAPIURL,
		client:   &http.Client{},
	}, nil
}

// pushoverPriority maps alert levels to Pushover priority values.
func pushoverPriority(level Level) int {
	switch level {
	case LevelCritical:
		return 1 // high priority
	case LevelWarning:
		return 0 // normal priority
	default:
		return -1 // low priority
	}
}

// Send delivers the alert via Pushover.
func (n *PushoverNotifier) Send(a Alert) error {
	payload := map[string]interface{}{
		"token":    n.appToken,
		"user":     n.userKey,
		"title":    fmt.Sprintf("VaultWatch – %s", a.Level),
		"message":  a.String(),
		"priority": pushoverPriority(a.Level),
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("pushover: marshal payload: %w", err)
	}

	resp, err := n.client.Post(n.apiURL, "application/json", strings.NewReader(string(body)))
	if err != nil {
		return fmt.Errorf("pushover: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("pushover: unexpected status %d", resp.StatusCode)
	}
	return nil
}
