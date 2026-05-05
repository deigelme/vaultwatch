package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeDatadog(t *testing.T, statusCode int, fn func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewDatadogNotifier_EmptyKey(t *testing.T) {
	_, err := NewDatadogNotifier("", "")
	if err == nil {
		t.Fatal("expected error for empty API key, got nil")
	}
}

func TestDatadogNotifier_Send_Success(t *testing.T) {
	var gotAPIKey string
	var gotEvent map[string]interface{}

	srv := startFakeDatadog(t, http.StatusAccepted, func(r *http.Request) {
		gotAPIKey = r.Header.Get("DD-API-KEY")
		if err := json.NewDecoder(r.Body).Decode(&gotEvent); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
	})
	defer srv.Close()

	n, err := NewDatadogNotifier("test-api-key", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/my-service/db",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotAPIKey != "test-api-key" {
		t.Errorf("expected DD-API-KEY header 'test-api-key', got %q", gotAPIKey)
	}
	if gotEvent["alert_type"] != "warning" {
		t.Errorf("expected alert_type 'warning', got %v", gotEvent["alert_type"])
	}
}

func TestDatadogNotifier_Send_CriticalAlertType(t *testing.T) {
	var gotEvent map[string]interface{}

	srv := startFakeDatadog(t, http.StatusAccepted, func(r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotEvent)
	})
	defer srv.Close()

	n, _ := NewDatadogNotifier("key", srv.URL)
	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/prod/token",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
		TimeLeft:   2 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotEvent["alert_type"] != "error" {
		t.Errorf("expected alert_type 'error' for critical, got %v", gotEvent["alert_type"])
	}
}

func TestDatadogNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeDatadog(t, http.StatusForbidden, nil)
	defer srv.Close()

	n, _ := NewDatadogNotifier("bad-key", srv.URL)
	a := Alert{
		Level:      LevelInfo,
		SecretPath: "secret/test",
		ExpiresAt:  time.Now().Add(72 * time.Hour),
		TimeLeft:   72 * time.Hour,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
