package alert

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// ZulipNotifier sends alerts to a Zulip stream via the Zulip REST API.
type ZulipNotifier struct {
	baseURL string
	bot     string
	apiKey  string
	stream  string
	topic   string
	client  *http.Client
}

// NewZulipNotifier creates a ZulipNotifier.
// baseURL is the Zulip server URL (e.g. "https://yourorg.zulipchat.com"),
// bot is the bot email address, apiKey is the bot API key,
// stream and topic identify where messages are posted.
func NewZulipNotifier(baseURL, bot, apiKey, stream, topic string) (*ZulipNotifier, error) {
	if strings.TrimSpace(baseURL) == "" {
		return nil, fmt.Errorf("zulip: base URL must not be empty")
	}
	if strings.TrimSpace(bot) == "" {
		return nil, fmt.Errorf("zulip: bot email must not be empty")
	}
	if strings.TrimSpace(apiKey) == "" {
		return nil, fmt.Errorf("zulip: API key must not be empty")
	}
	if strings.TrimSpace(stream) == "" {
		return nil, fmt.Errorf("zulip: stream must not be empty")
	}
	return &ZulipNotifier{
		baseURL: strings.TrimRight(baseURL, "/"),
		bot:     bot,
		apiKey:  apiKey,
		stream:  stream,
		topic:   topic,
		client:  &http.Client{},
	}, nil
}

// Send posts an alert to the configured Zulip stream and topic.
func (z *ZulipNotifier) Send(a Alert) error {
	endpoint := z.baseURL + "/api/v1/messages"

	form := url.Values{}
	form.Set("type", "stream")
	form.Set("to", z.stream)
	form.Set("topic", z.topic)
	form.Set("content", fmt.Sprintf("**[%s]** %s", a.Level, a.String()))

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return fmt.Errorf("zulip: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth(z.bot, z.apiKey)

	resp, err := z.client.Do(req)
	if err != nil {
		return fmt.Errorf("zulip: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("zulip: unexpected status %d", resp.StatusCode)
	}
	return nil
}
