package alert

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeNtfy(t *testing.T, wantPriority string, handler func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(r)
		}
		w.WriteHeader(http.StatusOK)
	}))
}

func TestNewNtfyNotifier_EmptyBaseURL(t *testing.T) {
	_, err := NewNtfyNotifier("", "alerts")
	if err == nil {
		t.Fatal("expected error for empty baseURL, got nil")
	}
}

func TestNewNtfyNotifier_EmptyTopic(t *testing.T) {
	_, err := NewNtfyNotifier("https://ntfy.sh", "")
	if err == nil {
		t.Fatal("expected error for empty topic, got nil")
	}
}

func TestNtfyNotifier_Send_Success(t *testing.T) {
	var gotPath, gotPriority, gotTitle string
	var gotBody []byte

	srv := startFakeNtfy(t, "high", func(r *http.Request) {
		gotPath = r.URL.Path
		gotPriority = r.Header.Get("Priority")
		gotTitle = r.Header.Get("Title")
		gotBody, _ = io.ReadAll(r.Body)
	})
	defer srv.Close()

	n, err := NewNtfyNotifier(srv.URL, "vaultwatch")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db/prod",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Level:      LevelWarning,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if gotPath != "/vaultwatch" {
		t.Errorf("expected path /vaultwatch, got %s", gotPath)
	}
	if gotPriority != "high" {
		t.Errorf("expected priority 'high', got %s", gotPriority)
	}
	if gotTitle == "" {
		t.Error("expected non-empty Title header")
	}
	if len(gotBody) == 0 {
		t.Error("expected non-empty body")
	}
}

func TestNtfyNotifier_Send_CriticalPriority(t *testing.T) {
	var gotPriority string
	srv := startFakeNtfy(t, "urgent", func(r *http.Request) {
		gotPriority = r.Header.Get("Priority")
	})
	defer srv.Close()

	n, _ := NewNtfyNotifier(srv.URL, "vaultwatch")
	a := Alert{
		SecretPath: "secret/api/key",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
		TimeLeft:   2 * time.Hour,
		Level:      LevelCritical,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotPriority != "urgent" {
		t.Errorf("expected priority 'urgent', got %s", gotPriority)
	}
}

func TestNtfyNotifier_Send_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	n, _ := NewNtfyNotifier(srv.URL, "vaultwatch")
	a := Alert{
		SecretPath: "secret/db/prod",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		TimeLeft:   24 * time.Hour,
		Level:      LevelWarning,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status, got nil")
	}
}
