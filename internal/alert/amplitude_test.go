package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeAmplitude(t *testing.T, statusCode int, validate func(body []byte)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if validate != nil {
			validate(body)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewAmplitudeNotifier_EmptyKey(t *testing.T) {
	_, err := NewAmplitudeNotifier("", "")
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
}

func TestAmplitudeNotifier_Send_Success(t *testing.T) {
	var capturedBody []byte
	srv := startFakeAmplitude(t, http.StatusOK, func(b []byte) {
		capturedBody = b
	})
	defer srv.Close()

	n, err := NewAmplitudeNotifier("test-api-key", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db/password",
		Level:      LevelWarning,
		Message:    "Secret expiring soon",
		TimeLeft:   48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(capturedBody, &payload); err != nil {
		t.Fatalf("invalid JSON payload: %v", err)
	}
	if payload["api_key"] != "test-api-key" {
		t.Errorf("expected api_key=test-api-key, got %v", payload["api_key"])
	}
	events, ok := payload["events"].([]interface{})
	if !ok || len(events) != 1 {
		t.Fatalf("expected 1 event, got %v", payload["events"])
	}
	ev := events[0].(map[string]interface{})
	if ev["event_type"] != "secret_expiry_alert" {
		t.Errorf("unexpected event_type: %v", ev["event_type"])
	}
	props := ev["event_properties"].(map[string]interface{})
	if props["secret_path"] != "secret/db/password" {
		t.Errorf("unexpected secret_path: %v", props["secret_path"])
	}
}

func TestAmplitudeNotifier_Send_DefaultEndpoint(t *testing.T) {
	// Verify that an empty endpoint falls back to the default (no error on construction).
	n, err := NewAmplitudeNotifier("key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.endpoint != amplitudeDefaultEndpoint {
		t.Errorf("expected default endpoint %q, got %q", amplitudeDefaultEndpoint, n.endpoint)
	}
}

func TestAmplitudeNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeAmplitude(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, err := NewAmplitudeNotifier("key", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/api/token",
		Level:      LevelCritical,
		Message:    "Critical: secret expired",
		TimeLeft:   1 * time.Hour,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
