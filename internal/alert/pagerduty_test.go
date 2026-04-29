package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakePagerDuty(t *testing.T, statusCode int, capturedBody *[]byte) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		*capturedBody = body
		w.WriteHeader(statusCode)
	}))
}

func TestNewPagerDutyNotifier_EmptyKey(t *testing.T) {
	_, err := NewPagerDutyNotifier("")
	if err == nil {
		t.Fatal("expected error for empty integration key, got nil")
	}
}

func TestPagerDutyNotifier_Send_Success(t *testing.T) {
	var captured []byte
	srv := startFakePagerDuty(t, http.StatusAccepted, &captured)
	defer srv.Close()

	n, err := NewPagerDutyNotifier("test-key-123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.eventURL = srv.URL

	a := Alert{
		SecretPath: "secret/db/password",
		ExpiresAt:  time.Now().Add(12 * time.Hour),
		TimeLeft:   12 * time.Hour,
		Level:      Warning,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(captured, &payload); err != nil {
		t.Fatalf("could not parse request body: %v", err)
	}
	if payload["routing_key"] != "test-key-123" {
		t.Errorf("expected routing_key 'test-key-123', got %v", payload["routing_key"])
	}
	if payload["event_action"] != "trigger" {
		t.Errorf("expected event_action 'trigger', got %v", payload["event_action"])
	}
	p := payload["payload"].(map[string]interface{})
	if p["severity"] != "warning" {
		t.Errorf("expected severity 'warning', got %v", p["severity"])
	}
}

func TestPagerDutyNotifier_Send_CriticalSeverity(t *testing.T) {
	var captured []byte
	srv := startFakePagerDuty(t, http.StatusAccepted, &captured)
	defer srv.Close()

	n, _ := NewPagerDutyNotifier("key")
	n.eventURL = srv.URL

	a := Alert{
		SecretPath: "secret/api/token",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
		TimeLeft:   2 * time.Hour,
		Level:      Critical,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	var payload map[string]interface{}
	json.Unmarshal(captured, &payload)
	p := payload["payload"].(map[string]interface{})
	if p["severity"] != "critical" {
		t.Errorf("expected severity 'critical', got %v", p["severity"])
	}
}

func TestPagerDutyNotifier_Send_NonOKStatus(t *testing.T) {
	var captured []byte
	srv := startFakePagerDuty(t, http.StatusBadRequest, &captured)
	defer srv.Close()

	n, _ := NewPagerDutyNotifier("key")
	n.eventURL = srv.URL

	a := Alert{SecretPath: "secret/x", Level: Warning, TimeLeft: time.Hour}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
