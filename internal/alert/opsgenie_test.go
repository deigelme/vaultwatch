package alert

import (
	"encoding/json"
	"io"
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
		t.Fatal("expected error for empty api key")
	}
}

func TestOpsGenieNotifier_Send_Success(t *testing.T) {
	var gotAuth, gotContentType string
	var gotBody map[string]interface{}

	srv := startFakeOpsGenie(t, 202, func(r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
	})
	defer srv.Close()

	n, err := NewOpsGenieNotifier("test-key", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/myapp/db",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if gotAuth != "GenieKey test-key" {
		t.Errorf("expected GenieKey auth, got %q", gotAuth)
	}
	if gotContentType != "application/json" {
		t.Errorf("unexpected content-type: %q", gotContentType)
	}
	if gotBody["priority"] != "P3" {
		t.Errorf("expected P3 priority, got %v", gotBody["priority"])
	}
}

func TestOpsGenieNotifier_Send_CriticalPriority(t *testing.T) {
	var gotBody map[string]interface{}
	srv := startFakeOpsGenie(t, 202, func(r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
	})
	defer srv.Close()

	n, _ := NewOpsGenieNotifier("key", srv.URL)
	a := Alert{Level: LevelCritical, SecretPath: "secret/crit", ExpiresAt: time.Now().Add(time.Hour)}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if gotBody["priority"] != "P1" {
		t.Errorf("expected P1 for critical, got %v", gotBody["priority"])
	}
}

func TestOpsGenieNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeOpsGenie(t, 429, nil)
	defer srv.Close()

	n, _ := NewOpsGenieNotifier("key", srv.URL)
	a := Alert{Level: LevelInfo, SecretPath: "secret/x", ExpiresAt: time.Now().Add(72 * time.Hour)}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}

func TestOpsGenieNotifier_Send_DefaultEndpoint(t *testing.T) {
	n, err := NewOpsGenieNotifier("mykey", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.endpoint != "https://api.opsgenie.com/v2/alerts" {
		t.Errorf("unexpected default endpoint: %s", n.endpoint)
	}
}
