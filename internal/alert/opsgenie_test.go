package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeOpsGenie(t *testing.T, statusCode int, handler func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewOpsGenieNotifier_EmptyKey(t *testing.T) {
	_, err := NewOpsGenieNotifier("", "")
	if err == nil {
		t.Fatal("expected error for empty api key, got nil")
	}
}

func TestOpsGenieNotifier_Send_Success(t *testing.T) {
	var gotAuth, gotContentType string
	var gotPayload opsGeniePayload

	srv := startFakeOpsGenie(t, http.StatusAccepted, func(r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Errorf("decode body: %v", err)
		}
	})
	defer srv.Close()

	n, err := NewOpsGenieNotifier("test-key-123", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/my-app/db",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if gotAuth != "GenieKey test-key-123" {
		t.Errorf("expected Authorization header 'GenieKey test-key-123', got %q", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", gotContentType)
	}
	if gotPayload.Priority != "P2" {
		t.Errorf("expected priority P2 for warning, got %q", gotPayload.Priority)
	}
	if gotPayload.Details["secret_path"] != "secret/my-app/db" {
		t.Errorf("unexpected secret_path in details: %q", gotPayload.Details["secret_path"])
	}
}

func TestOpsGenieNotifier_Send_CriticalPriority(t *testing.T) {
	var gotPayload opsGeniePayload
	srv := startFakeOpsGenie(t, http.StatusAccepted, func(r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotPayload)
	})
	defer srv.Close()

	n, _ := NewOpsGenieNotifier("key", srv.URL)
	a := Alert{Level: LevelCritical, SecretPath: "secret/crit", ExpiresAt: time.Now().Add(time.Hour), TimeLeft: time.Hour}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotPayload.Priority != "P1" {
		t.Errorf("expected P1 for critical, got %q", gotPayload.Priority)
	}
}

func TestOpsGenieNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeOpsGenie(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	n, _ := NewOpsGenieNotifier("bad-key", srv.URL)
	a := Alert{Level: LevelInfo, SecretPath: "secret/x", ExpiresAt: time.Now().Add(72 * time.Hour), TimeLeft: 72 * time.Hour}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
