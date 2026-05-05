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

func startFakeGoogleChat(t *testing.T, statusCode int, handler func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewGoogleChatNotifier_EmptyURL(t *testing.T) {
	_, err := NewGoogleChatNotifier("")
	if err == nil {
		t.Fatal("expected error for empty webhook URL, got nil")
	}
}

func TestGoogleChatNotifier_Send_Success(t *testing.T) {
	var gotBody []byte
	srv := startFakeGoogleChat(t, http.StatusOK, func(r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
	})
	defer srv.Close()

	n, err := NewGoogleChatNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     LevelWarning,
		Message:   "secret/db expires in 6 days",
		Secret:    "secret/db",
		ExpiresAt: time.Now().Add(6 * 24 * time.Hour),
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("failed to unmarshal request body: %v", err)
	}
	if !strings.Contains(payload["text"], "VaultWatch Alert") {
		t.Errorf("expected 'VaultWatch Alert' in text, got: %s", payload["text"])
	}
	if !strings.Contains(payload["text"], string(LevelWarning)) {
		t.Errorf("expected level %q in text, got: %s", LevelWarning, payload["text"])
	}
}

func TestGoogleChatNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeGoogleChat(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, err := NewGoogleChatNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:   LevelCritical,
		Message: "secret/api expires in 1 day",
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}
