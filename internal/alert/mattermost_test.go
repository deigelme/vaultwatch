package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeMattermost(t *testing.T, statusCode int, gotPayload *mattermostPayload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotPayload != nil {
			if err := json.NewDecoder(r.Body).Decode(gotPayload); err != nil {
				t.Errorf("failed to decode payload: %v", err)
			}
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewMattermostNotifier_EmptyURL(t *testing.T) {
	_, err := NewMattermostNotifier("", "#alerts", "vaultwatch")
	if err == nil {
		t.Fatal("expected error for empty webhook URL")
	}
}

func TestMattermostNotifier_Send_Success(t *testing.T) {
	var got mattermostPayload
	srv := startFakeMattermost(t, http.StatusOK, &got)
	defer srv.Close()

	n, err := NewMattermostNotifier(srv.URL, "#vault-alerts", "vaultwatch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     Warning,
		Secret:    "secret/db/password",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		TimeLeft:  48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	if got.Channel != "#vault-alerts" {
		t.Errorf("expected channel #vault-alerts, got %q", got.Channel)
	}
	if got.Username != "vaultwatch" {
		t.Errorf("expected username vaultwatch, got %q", got.Username)
	}
	if got.Text == "" {
		t.Error("expected non-empty text")
	}
}

func TestMattermostNotifier_Send_CriticalEmoji(t *testing.T) {
	var got mattermostPayload
	srv := startFakeMattermost(t, http.StatusOK, &got)
	defer srv.Close()

	n, err := NewMattermostNotifier(srv.URL, "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     Critical,
		Secret:    "secret/api/key",
		ExpiresAt: time.Now().Add(2 * time.Hour),
		TimeLeft:  2 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	if got.IconEmoji != ":rotating_light:" {
		t.Errorf("expected :rotating_light: for critical, got %q", got.IconEmoji)
	}
}

func TestMattermostNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeMattermost(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, err := NewMattermostNotifier(srv.URL, "#alerts", "bot")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:    Warning,
		Secret:   "secret/test",
		TimeLeft: 24 * time.Hour,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
