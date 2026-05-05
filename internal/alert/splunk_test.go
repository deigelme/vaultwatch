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

func startFakeSplunk(t *testing.T, statusCode int, capturedReq *splunkEvent) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capturedReq != nil {
			body, _ := io.ReadAll(r.Body)
			_ = json.Unmarshal(body, capturedReq)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewSplunkNotifier_EmptyURL(t *testing.T) {
	_, err := NewSplunkNotifier("", "token")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewSplunkNotifier_EmptyToken(t *testing.T) {
	_, err := NewSplunkNotifier("http://splunk.example.com", "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestSplunkNotifier_Send_Success(t *testing.T) {
	var captured splunkEvent
	srv := startFakeSplunk(t, http.StatusOK, &captured)
	defer srv.Close()

	n, err := NewSplunkNotifier(srv.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		Message:    "secret expiring soon",
		SecretPath: "secret/db/password",
		TimeLeft:   48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	if captured.Source != "vaultwatch" {
		t.Errorf("expected source 'vaultwatch', got %q", captured.Source)
	}
	if captured.Event.Secret != a.SecretPath {
		t.Errorf("expected secret path %q, got %q", a.SecretPath, captured.Event.Secret)
	}
	if !strings.Contains(captured.Event.Severity, string(LevelWarning)) {
		t.Errorf("expected severity %q in payload, got %q", LevelWarning, captured.Event.Severity)
	}
}

func TestSplunkNotifier_Send_CriticalSeverity(t *testing.T) {
	var captured splunkEvent
	srv := startFakeSplunk(t, http.StatusOK, &captured)
	defer srv.Close()

	n, _ := NewSplunkNotifier(srv.URL, "test-token")
	a := Alert{
		Level:      LevelCritical,
		Message:    "secret critically close to expiry",
		SecretPath: "secret/api/key",
		TimeLeft:   2 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if captured.Event.Severity != string(LevelCritical) {
		t.Errorf("expected critical severity, got %q", captured.Event.Severity)
	}
}

func TestSplunkNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeSplunk(t, http.StatusForbidden, nil)
	defer srv.Close()

	n, _ := NewSplunkNotifier(srv.URL, "bad-token")
	err := n.Send(Alert{Level: LevelWarning, Message: "test", SecretPath: "s", TimeLeft: time.Hour})
	if err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
