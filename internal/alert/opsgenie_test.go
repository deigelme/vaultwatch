package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeOpsGenie(t *testing.T, statusCode int, fn func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewOpsGenieNotifier_EmptyKey(t *testing.T) {
	_, err := NewOpsGenieNotifier("", "")
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
}

func TestOpsGenieNotifier_Send_Success(t *testing.T) {
	var gotAuth string
	var gotPayload opsGeniePayload

	srv := startFakeOpsGenie(t, http.StatusAccepted, func(r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Errorf("decode body: %v", err)
		}
	})
	defer srv.Close()

	n, err := NewOpsGenieNotifier("test-api-key", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/my-service/db",
		TimeLeft:   48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	if gotAuth != "GenieKey test-api-key" {
		t.Errorf("auth header = %q, want %q", gotAuth, "GenieKey test-api-key")
	}
	if gotPayload.Priority != "P3" {
		t.Errorf("priority = %q, want P3", gotPayload.Priority)
	}
	if gotPayload.Details["secret_path"] != "secret/my-service/db" {
		t.Errorf("details secret_path = %q", gotPayload.Details["secret_path"])
	}
}

func TestOpsGenieNotifier_Send_CriticalPriority(t *testing.T) {
	var gotPayload opsGeniePayload
	srv := startFakeOpsGenie(t, http.StatusAccepted, func(r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&gotPayload)
	})
	defer srv.Close()

	n, _ := NewOpsGenieNotifier("key", srv.URL)
	a := Alert{Level: LevelCritical, SecretPath: "secret/crit", TimeLeft: 1 * time.Hour}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if gotPayload.Priority != "P1" {
		t.Errorf("priority = %q, want P1", gotPayload.Priority)
	}
}

func TestOpsGenieNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeOpsGenie(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	n, _ := NewOpsGenieNotifier("bad-key", srv.URL)
	a := Alert{Level: LevelWarning, SecretPath: "secret/x", TimeLeft: 24 * time.Hour}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
