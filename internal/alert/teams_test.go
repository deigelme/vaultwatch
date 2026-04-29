package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeTeams(t *testing.T, statusCode int, gotPayload *teamsPayload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, gotPayload)
		w.WriteHeader(statusCode)
	}))
}

func TestNewTeamsNotifier_EmptyURL(t *testing.T) {
	_, err := NewTeamsNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestTeamsNotifier_Send_Success(t *testing.T) {
	var got teamsPayload
	srv := startFakeTeams(t, http.StatusOK, &got)
	defer srv.Close()

	n, err := NewTeamsNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/myapp/db",
		Expiry:     time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if got.Type != "message" {
		t.Errorf("expected type 'message', got %q", got.Type)
	}
	if len(got.Attachments) == 0 {
		t.Fatal("expected at least one attachment")
	}
	body := got.Attachments[0].Content.Body
	if len(body) < 2 {
		t.Fatalf("expected at least 2 body elements, got %d", len(body))
	}
	if body[0].Text != "VaultWatch Alert" {
		t.Errorf("unexpected header text: %q", body[0].Text)
	}
}

func TestTeamsNotifier_Send_NonOKStatus(t *testing.T) {
	var got teamsPayload
	srv := startFakeTeams(t, http.StatusInternalServerError, &got)
	defer srv.Close()

	n, err := NewTeamsNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/myapp/token",
		Expiry:     time.Now().Add(2 * time.Hour),
		TimeLeft:   2 * time.Hour,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}
