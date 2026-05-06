package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startFakeLark(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", ct)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewLarkNotifier_EmptyURL(t *testing.T) {
	_, err := NewLarkNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestLarkNotifier_Send_Success(t *testing.T) {
	server := startFakeLark(t, http.StatusOK)
	defer server.Close()

	n, err := NewLarkNotifier(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/my-service/api-key",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}
}

func TestLarkNotifier_Send_PayloadContent(t *testing.T) {
	var received larkBody

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	n, _ := NewLarkNotifier(server.URL)
	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/db/password",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
		TimeLeft:   2 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	if received.MsgType != "text" {
		t.Errorf("expected msg_type=text, got %s", received.MsgType)
	}
	if !strings.Contains(received.Content.Text, "secret/db/password") {
		t.Errorf("expected secret path in message text, got: %s", received.Content.Text)
	}
	if !strings.Contains(received.Content.Text, string(LevelCritical)) {
		t.Errorf("expected alert level in message text, got: %s", received.Content.Text)
	}
}

func TestLarkNotifier_Send_NonOKStatus(t *testing.T) {
	server := startFakeLark(t, http.StatusInternalServerError)
	defer server.Close()

	n, _ := NewLarkNotifier(server.URL)
	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/test",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		TimeLeft:   24 * time.Hour,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}
