package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeNewRelic(t *testing.T, statusCode int, captured *[]newRelicEvent) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if captured != nil {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, captured)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewNewRelicNotifier_EmptyAccountID(t *testing.T) {
	_, err := NewNewRelicNotifier("", "apikey")
	if err == nil {
		t.Fatal("expected error for empty account ID")
	}
}

func TestNewNewRelicNotifier_EmptyAPIKey(t *testing.T) {
	_, err := NewNewRelicNotifier("12345", "")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestNewRelicNotifier_Send_Success(t *testing.T) {
	var captured []newRelicEvent
	srv := startFakeNewRelic(t, http.StatusOK, &captured)
	defer srv.Close()

	n, err := NewNewRelicNotifier("99999", "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Override URL to point at fake server.
	n.url = srv.URL

	a := Alert{
		Level:      LevelWarning,
		Message:    "secret expiring soon",
		SecretPath: "secret/db/password",
		TimeLeft:   36 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	if len(captured) != 1 {
		t.Fatalf("expected 1 event, got %d", len(captured))
	}
	if captured[0].EventType != "VaultWatchAlert" {
		t.Errorf("expected eventType 'VaultWatchAlert', got %q", captured[0].EventType)
	}
	if captured[0].SecretPath != a.SecretPath {
		t.Errorf("expected secret path %q, got %q", a.SecretPath, captured[0].SecretPath)
	}
	if captured[0].Severity != string(LevelWarning) {
		t.Errorf("expected severity %q, got %q", LevelWarning, captured[0].Severity)
	}
}

func TestNewRelicNotifier_Send_CriticalSeverity(t *testing.T) {
	var captured []newRelicEvent
	srv := startFakeNewRelic(t, http.StatusOK, &captured)
	defer srv.Close()

	n, _ := NewNewRelicNotifier("99999", "test-key")
	n.url = srv.URL

	a := Alert{
		Level:      LevelCritical,
		Message:    "secret critically close to expiry",
		SecretPath: "secret/api/key",
		TimeLeft:   1 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if len(captured) == 0 || captured[0].Severity != string(LevelCritical) {
		t.Errorf("expected critical severity in event")
	}
}

func TestNewRelicNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeNewRelic(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	n, _ := NewNewRelicNotifier("99999", "bad-key")
	n.url = srv.URL

	err := n.Send(Alert{Level: LevelWarning, Message: "test", SecretPath: "s", TimeLeft: time.Hour})
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
