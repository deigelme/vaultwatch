package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeXMPP(t *testing.T, statusCode int, gotBody *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		*gotBody = string(raw)
		w.WriteHeader(statusCode)
	}))
}

func TestNewXMPPNotifier_EmptyGatewayURL(t *testing.T) {
	_, err := NewXMPPNotifier("", "user@example.com")
	if err == nil {
		t.Fatal("expected error for empty gateway URL")
	}
}

func TestNewXMPPNotifier_EmptyRecipient(t *testing.T) {
	_, err := NewXMPPNotifier("http://localhost:9999", "")
	if err == nil {
		t.Fatal("expected error for empty recipient JID")
	}
}

func TestXMPPNotifier_Send_Success(t *testing.T) {
	var body string
	srv := startFakeXMPP(t, http.StatusOK, &body)
	defer srv.Close()

	n, err := NewXMPPNotifier(srv.URL, "ops@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Path:      "secret/db",
		Level:     LevelWarning,
		ExpiresAt: time.Now().Add(48 * time.Hour),
		TimeLeft:  48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("could not parse request body as JSON: %v", err)
	}
	if payload["to"] != "ops@example.com" {
		t.Errorf("expected to=ops@example.com, got %q", payload["to"])
	}
	if payload["body"] == "" {
		t.Error("expected non-empty body field")
	}
}

func TestXMPPNotifier_Send_NonOKStatus(t *testing.T) {
	var body string
	srv := startFakeXMPP(t, http.StatusInternalServerError, &body)
	defer srv.Close()

	n, err := NewXMPPNotifier(srv.URL, "ops@example.com")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Path:      "secret/db",
		Level:     LevelCritical,
		ExpiresAt: time.Now().Add(2 * time.Hour),
		TimeLeft:  2 * time.Hour,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
