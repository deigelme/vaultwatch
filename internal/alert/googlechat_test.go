package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startFakeGoogleChat(t *testing.T, statusCode int, fn func(body []byte)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if fn != nil {
			fn(body)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewGoogleChatNotifier_EmptyURL(t *testing.T) {
	_, err := NewGoogleChatNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestGoogleChatNotifier_Send_Success(t *testing.T) {
	var received []byte
	srv := startFakeGoogleChat(t, http.StatusOK, func(body []byte) {
		received = body
	})
	defer srv.Close()

	n, err := NewGoogleChatNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/my-app/db",
		Expiry:     time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Message:    "expires soon",
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(received, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}
	if !strings.Contains(payload["text"], "secret/my-app/db") {
		t.Errorf("payload text missing secret path: %s", payload["text"])
	}
	if !strings.Contains(payload["text"], "WARNING") {
		t.Errorf("payload text missing level: %s", payload["text"])
	}
}

func TestGoogleChatNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeGoogleChat(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, _ := NewGoogleChatNotifier(srv.URL)
	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/creds",
		Expiry:     time.Now().Add(1 * time.Hour),
		TimeLeft:   1 * time.Hour,
		Message:    "critical expiry",
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-200 status")
	}
}
