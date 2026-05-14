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

func startFakeGoogleChat(t *testing.T, statusCode int, fn func(body map[string]string)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		var payload map[string]string
		_ = json.Unmarshal(b, &payload)
		if fn != nil {
			fn(payload)
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
	var got map[string]string
	srv := startFakeGoogleChat(t, http.StatusOK, func(body map[string]string) {
		got = body
	})
	defer srv.Close()

	n, err := NewGoogleChatNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     LevelWarning,
		Path:      "secret/my-app/db",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		TimeLeft:  48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if !strings.Contains(got["text"], "VaultWatch Alert") {
		t.Errorf("expected alert text, got: %q", got["text"])
	}
	if !strings.Contains(got["text"], "secret/my-app/db") {
		t.Errorf("expected path in text, got: %q", got["text"])
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
		Level:     LevelCritical,
		Path:      "secret/prod/cert",
		ExpiresAt: time.Now().Add(2 * time.Hour),
		TimeLeft:  2 * time.Hour,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
