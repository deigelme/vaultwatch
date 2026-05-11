package alert

import (
	"fmt"
	"net/http"
	"strings"
)

// XMPPNotifier sends alerts via an XMPP HTTP gateway (e.g. a self-hosted
// rest-xmpp bridge). It POSTs a JSON body to the gateway endpoint so that
// the gateway forwards the message to the configured XMPP JID.
type XMPPNotifier struct {
	gatewayURL string
	to         string
	client     *http.Client
}

// NewXMPPNotifier creates an XMPPNotifier that delivers messages through the
// given HTTP gateway URL to the recipient JID.
func NewXMPPNotifier(gatewayURL, to string) (*XMPPNotifier, error) {
	if strings.TrimSpace(gatewayURL) == "" {
		return nil, fmt.Errorf("xmpp: gateway URL must not be empty")
	}
	if strings.TrimSpace(to) == "" {
		return nil, fmt.Errorf("xmpp: recipient JID must not be empty")
	}
	return &XMPPNotifier{
		gatewayURL: gatewayURL,
		to:         to,
		client:     &http.Client{},
	}, nil
}

// Send delivers the alert through the XMPP HTTP gateway.
func (n *XMPPNotifier) Send(a Alert) error {
	body := fmt.Sprintf(
		`{"to":%q,"body":%q}`,
		n.to,
		a.String(),
	)
	resp, err := n.client.Post(
		n.gatewayURL,
		"application/json",
		strings.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("xmpp: gateway request failed: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("xmpp: gateway returned non-2xx status %d", resp.StatusCode)
	}
	return nil
}
