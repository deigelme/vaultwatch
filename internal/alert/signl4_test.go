package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeSIGNL4(t *testing.T, statusCode int, capture *signl4Payload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capture != nil {
			if err := json.NewDecoder(r.Body).Decode(capture); err != nil {
				t.Errorf("signl4: failed to decode request body: %v", err)
			}
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewSIGNL4Notifier_EmptyURL(t *testing.T) {
	_, err := NewSIGNL4Notifier("")
	if err == nil {
		t.Fatal("expected error for empty webhook URL, got nil")
	}
}

func TestSIGNL4Notifier_Send_Success(t *testing.T) {
	var captured signl4Payload
	srv := startFakeSIGNL4(t, http.StatusOK, &captured)
	defer srv.Close()

	n, err := NewSIGNL4Notifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db/password",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		Level:      LevelWarning,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}

	if captured.Source != "vaultwatch" {
		t.Errorf("expected source 'vaultwatch', got %q", captured.Source)
	}
	if captured.Severity != 2 {
		t.Errorf("expected severity 2 for warning, got %d", captured.Severity)
	}
}

func TestSIGNL4Notifier_Send_CriticalSeverity(t *testing.T) {
	var captured signl4Payload
	srv := startFakeSIGNL4(t, http.StatusOK, &captured)
	defer srv.Close()

	n, err := NewSIGNL4Notifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/api/key",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
		Level:      LevelCritical,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}

	if captured.Severity != 3 {
		t.Errorf("expected severity 3 for critical, got %d", captured.Severity)
	}
}

func TestSIGNL4Notifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeSIGNL4(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, err := NewSIGNL4Notifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/token",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		Level:      LevelWarning,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}
