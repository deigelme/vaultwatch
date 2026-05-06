package alert

import (
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// HTTPGetNotifier sends an alert by performing an HTTP GET request to a
// configured URL, optionally appending alert fields as query parameters.
type HTTPGetNotifier struct {
	baseURL string
	client  *http.Client
}

// NewHTTPGetNotifier creates a new HTTPGetNotifier.
// baseURL must be a non-empty, valid URL.
func NewHTTPGetNotifier(baseURL string) (*HTTPGetNotifier, error) {
	if baseURL == "" {
		return nil, fmt.Errorf("httpget: base URL must not be empty")
	}
	if _, err := url.ParseRequestURI(baseURL); err != nil {
		return nil, fmt.Errorf("httpget: invalid base URL: %w", err)
	}
	return &HTTPGetNotifier{
		baseURL: baseURL,
		client:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Send performs an HTTP GET request with alert details encoded as query
// parameters: path, level, and message.
func (n *HTTPGetNotifier) Send(a Alert) error {
	params := url.Values{}
	params.Set("path", a.SecretPath)
	params.Set("level", string(a.Level))
	params.Set("message", a.Message)

	full := n.baseURL + "?" + params.Encode()

	resp, err := n.client.Get(full) //nolint:noctx
	if err != nil {
		return fmt.Errorf("httpget: request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("httpget: unexpected status %d", resp.StatusCode)
	}
	return nil
}
