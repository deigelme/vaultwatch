package alert

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// MatrixNotifier sends alerts to a Matrix room via the Matrix HTTP API.
type MatrixNotifier struct {
	homeserver string
	token      string
	roomID     string
	client     *http.Client
}

type matrixTextPayload struct {
	MsgType       string `json:"msgtype"`
	Body          string `json:"body"`
	FormattedBody string `json:"formatted_body,omitempty"`
	Format        string `json:"format,omitempty"`
}

// NewMatrixNotifier creates a MatrixNotifier. homeserver should be the base URL
// (e.g. "https://matrix.org"), token is the access token, and roomID is the
// fully-qualified room ID (e.g. "!abc123:matrix.org").
func NewMatrixNotifier(homeserver, token, roomID string) (*MatrixNotifier, error) {
	if strings.TrimSpace(homeserver) == "" {
		return nil, fmt.Errorf("matrix: homeserver URL must not be empty")
	}
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("matrix: access token must not be empty")
	}
	if strings.TrimSpace(roomID) == "" {
		return nil, fmt.Errorf("matrix: room ID must not be empty")
	}
	return &MatrixNotifier{
		homeserver: strings.TrimRight(homeserver, "/"),
		token:      token,
		roomID:     roomID,
		client:     &http.Client{},
	}, nil
}

// Send posts an alert message to the configured Matrix room.
func (m *MatrixNotifier) Send(a Alert) error {
	payload := matrixTextPayload{
		MsgType:       "m.text",
		Body:          a.String(),
		FormattedBody: fmt.Sprintf("<b>[%s]</b> %s", a.Level, a.String()),
		Format:        "org.matrix.custom.html",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("matrix: marshal payload: %w", err)
	}

	url := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message",
		m.homeserver, m.roomID)

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("matrix: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.token)

	resp, err := m.client.Do(req)
	if err != nil {
		return fmt.Errorf("matrix: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("matrix: unexpected status %d", resp.StatusCode)
	}
	return nil
}
