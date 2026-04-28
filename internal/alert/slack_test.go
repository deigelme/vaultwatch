package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

func startFakeSlack(t *testing.T, statusCode int, received *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("slack: failed to decode payload: %v", err)
		}
		*received = payload["text"]
		w.WriteHeader(statusCode)
	}))
}

func TestNewSlackNotifier_EmptyURL(t *testing.T) {
	_, err := alert.NewSlackNotifier("")
	if err == nil {
		t.Fatal("expected error for empty webhook URL, got nil")
	}
}

func TestSlackNotifier_Send_Success(t *testing.T) {
	var received string
	srv := startFakeSlack(t, http.StatusOK, &received)
	defer srv.Close()

	notifier, err := alert.NewSlackNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := alert.Alert{
		SecretPath: "secret/db/password",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Level:      alert.LevelWarning,
	}

	if err := notifier.Send(a); err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}

	if received == "" {
		t.Error("expected non-empty message to be received by fake Slack server")
	}
}

func TestSlackNotifier_Send_NonOKStatus(t *testing.T) {
	var received string
	srv := startFakeSlack(t, http.StatusInternalServerError, &received)
	defer srv.Close()

	notifier, err := alert.NewSlackNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := alert.Alert{
		SecretPath: "secret/api/key",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		TimeLeft:   24 * time.Hour,
		Level:      alert.LevelCritical,
	}

	if err := notifier.Send(a); err == nil {
		t.Fatal("expected error for non-200 status, got nil")
	}
}
