package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeGrafana(t *testing.T, statusCode int, fn func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewGrafanaNotifier_EmptyURL(t *testing.T) {
	_, err := NewGrafanaNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestGrafanaNotifier_Send_Success(t *testing.T) {
	var gotPayload grafanaPayload
	server := startFakeGrafana(t, http.StatusOK, func(r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
	})
	defer server.Close()

	n, err := NewGrafanaNotifier(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/my-app/token",
		ExpiresIn:  48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}

	if gotPayload.State != "pending" {
		t.Errorf("expected state 'pending', got %q", gotPayload.State)
	}
	if gotPayload.Title == "" {
		t.Error("expected non-empty title")
	}
}

func TestGrafanaNotifier_Send_CriticalState(t *testing.T) {
	var gotPayload grafanaPayload
	server := startFakeGrafana(t, http.StatusOK, func(r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
	})
	defer server.Close()

	n, _ := NewGrafanaNotifier(server.URL)
	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/db/password",
		ExpiresIn:  2 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send() returned unexpected error: %v", err)
	}

	if gotPayload.State != "alerting" {
		t.Errorf("expected state 'alerting', got %q", gotPayload.State)
	}
}

func TestGrafanaNotifier_Send_NonOKStatus(t *testing.T) {
	server := startFakeGrafana(t, http.StatusInternalServerError, nil)
	defer server.Close()

	n, _ := NewGrafanaNotifier(server.URL)
	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/api/key",
		ExpiresIn:  24 * time.Hour,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}
