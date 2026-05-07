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
	var gotAuth, gotAlias string
	var gotPriority string

	srv := startFakeOpsGenie(t, http.StatusAccepted, func(r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		gotAlias, _ = body["alias"].(string)
		gotPriority, _ = body["priority"].(string)
	})
	defer srv.Close()

	n, err := NewOpsGenieNotifier("test-key", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db",
		Level:      LevelWarning,
		TimeLeft:   48 * time.Hour,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotAuth != "GenieKey test-key" {
		t.Errorf("auth header = %q, want %q", gotAuth, "GenieKey test-key")
	}
	if gotAlias != "secret/db" {
		t.Errorf("alias = %q, want %q", gotAlias, "secret/db")
	}
	if gotPriority != "P2" {
		t.Errorf("priority = %q, want P2", gotPriority)
	}
}

func TestOpsGenieNotifier_Send_CriticalPriority(t *testing.T) {
	var gotPriority string
	srv := startFakeOpsGenie(t, http.StatusAccepted, func(r *http.Request) {
		var body map[string]interface{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		gotPriority, _ = body["priority"].(string)
	})
	defer srv.Close()

	n, _ := NewOpsGenieNotifier("key", srv.URL)
	_ = n.Send(Alert{Level: LevelCritical, TimeLeft: time.Hour})

	if gotPriority != "P1" {
		t.Errorf("priority = %q, want P1", gotPriority)
	}
}

func TestOpsGenieNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeOpsGenie(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	n, _ := NewOpsGenieNotifier("bad-key", srv.URL)
	err := n.Send(Alert{Level: LevelWarning, TimeLeft: time.Hour})
	if err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestOpsGenieNotifier_Send_DefaultEndpoint(t *testing.T) {
	n, err := NewOpsGenieNotifier("key", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.endpoint != "https://api.opsgenie.com/v2/alerts" {
		t.Errorf("default endpoint = %q", n.endpoint)
	}
}
