package alert

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// SignalWireNotifier sends SMS alerts via the SignalWire REST API.
type SignalWireNotifier struct {
	spaceURL string
	projectID string
	apiToken string
	from string
	to string
	client *http.Client
}

// NewSignalWireNotifier creates a SignalWireNotifier. spaceURL is the full
// SignalWire space URL (e.g. https://example.signalwire.com), projectID and
// apiToken are the REST credentials, from/to are E.164 phone numbers.
func NewSignalWireNotifier(spaceURL, projectID, apiToken, from, to string) (*SignalWireNotifier, error) {
	if spaceURL == "" {
		return nil, fmt.Errorf("signalwire: spaceURL must not be empty")
	}
	if projectID == "" {
		return nil, fmt.Errorf("signalwire: projectID must not be empty")
	}
	if apiToken == "" {
		return nil, fmt.Errorf("signalwire: apiToken must not be empty")
	}
	if from == "" {
		return nil, fmt.Errorf("signalwire: from number must not be empty")
	}
	if to == "" {
		return nil, fmt.Errorf("signalwire: to number must not be empty")
	}
	return &SignalWireNotifier{
		spaceURL:  strings.TrimRight(spaceURL, "/"),
		projectID: projectID,
		apiToken:  apiToken,
		from:      from,
		to:        to,
		client:    &http.Client{},
	}, nil
}

// Send delivers the alert as an SMS message.
func (n *SignalWireNotifier) Send(a Alert) error {
	endpoint := fmt.Sprintf("%s/api/laml/2010-04-01/Accounts/%s/Messages.json",
		n.spaceURL, n.projectID)

	body := url.Values{}
	body.Set("From", n.from)
	body.Set("To", n.to)
	body.Set("Body", a.String())

	req, err := http.NewRequest(http.MethodPost, endpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return fmt.Errorf("signalwire: build request: %w", err)
	}
	req.SetBasicAuth(n.projectID, n.apiToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := n.client.Do(req)
	if err != nil {
		return fmt.Errorf("signalwire: send request: %w", err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("signalwire: unexpected status %d", resp.StatusCode)
	}
	return nil
}
