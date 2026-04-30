package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeVictorOps(t *testing.T, statusCode int, fn func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewVictorOpsNotifier_EmptyURL(t *testing.T) {
	_, err := NewVictorOpsNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestVictorOpsNotifier_Send_Success(t *testing.T) {
	var received victorOpsPayload

	ts := startFakeVictorOps(t, http.StatusOK, func(r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("failed to decode request body: %v", err)
		}
	})
	defer ts.Close()

	n, err := NewVictorOpsNotifier(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      Warning,
		SecretPath: "secret/my-service/db",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	if received.MessageType != "WARNING" {
		t.Errorf("expected message_type WARNING, got %s", received.MessageType)
	}
	if received.EntityID != "secret/my-service/db" {
		t.Errorf("expected entity_id secret/my-service/db, got %s", received.EntityID)
	}
}

func TestVictorOpsNotifier_Send_CriticalMessageType(t *testing.T) {
	var received victorOpsPayload

	ts := startFakeVictorOps(t, http.StatusOK, func(r *http.Request) {
		json.NewDecoder(r.Body).Decode(&received) //nolint:errcheck
	})
	defer ts.Close()

	n, _ := NewVictorOpsNotifier(ts.URL)

	a := Alert{
		Level:      Critical,
		SecretPath: "secret/prod/api-key",
		ExpiresAt:  time.Now().Add(6 * time.Hour),
		TimeLeft:   6 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	if received.MessageType != "CRITICAL" {
		t.Errorf("expected message_type CRITICAL, got %s", received.MessageType)
	}
}

func TestVictorOpsNotifier_Send_NonOKStatus(t *testing.T) {
	ts := startFakeVictorOps(t, http.StatusInternalServerError, nil)
	defer ts.Close()

	n, _ := NewVictorOpsNotifier(ts.URL)

	a := Alert{
		Level:      Warning,
		SecretPath: "secret/test",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		TimeLeft:   24 * time.Hour,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}
