package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeRocketChat(t *testing.T, statusCode int, captureBody *rocketChatPayload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if captureBody != nil {
			_ = json.NewDecoder(r.Body).Decode(captureBody)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewRocketChatNotifier_EmptyURL(t *testing.T) {
	_, err := NewRocketChatNotifier("")
	if err == nil {
		t.Fatal("expected error for empty webhook URL, got nil")
	}
}

func TestRocketChatNotifier_Send_Success(t *testing.T) {
	var captured rocketChatPayload
	srv := startFakeRocketChat(t, http.StatusOK, &captured)
	defer srv.Close()

	n, err := NewRocketChatNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/data/myapp/db",
		ExpiresIn:  48 * time.Hour,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}

	if len(captured.Attachments) == 0 {
		t.Fatal("expected at least one attachment in payload")
	}
	if captured.Attachments[0].Title != "secret/data/myapp/db" {
		t.Errorf("expected attachment title %q, got %q", "secret/data/myapp/db", captured.Attachments[0].Title)
	}
	if captured.Attachments[0].Color != "#FFA500" {
		t.Errorf("expected warning color #FFA500, got %q", captured.Attachments[0].Color)
	}
}

func TestRocketChatNotifier_Send_CriticalColor(t *testing.T) {
	var captured rocketChatPayload
	srv := startFakeRocketChat(t, http.StatusOK, &captured)
	defer srv.Close()

	n, _ := NewRocketChatNotifier(srv.URL)
	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/data/myapp/token",
		ExpiresIn:  2 * time.Hour,
		ExpiresAt:  time.Now().Add(2 * time.Hour),
	}

	_ = n.Send(a)

	if len(captured.Attachments) == 0 {
		t.Fatal("expected attachment")
	}
	if captured.Attachments[0].Color != "#FF0000" {
		t.Errorf("expected critical color #FF0000, got %q", captured.Attachments[0].Color)
	}
}

func TestRocketChatNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeRocketChat(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, _ := NewRocketChatNotifier(srv.URL)
	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/data/myapp/key",
		ExpiresIn:  24 * time.Hour,
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}
