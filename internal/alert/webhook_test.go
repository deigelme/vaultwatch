package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeWebhook(t *testing.T, statusCode int) (*httptest.Server, *[]webhookPayload) {
	t.Helper()
	received := &[]webhookPayload{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var p webhookPayload
		if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
			t.Errorf("failed to decode webhook payload: %v", err)
		}
		*received = append(*received, p)
		w.WriteHeader(statusCode)
	}))
	t.Cleanup(server.Close)
	return server, received
}

func TestNewWebhookNotifier_EmptyURL(t *testing.T) {
	_, err := NewWebhookNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestWebhookNotifier_Send_Success(t *testing.T) {
	server, received := startFakeWebhook(t, http.StatusOK)

	n, err := NewWebhookNotifier(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/my-app/db",
		TimeLeft:   48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	if len(*received) != 1 {
		t.Fatalf("expected 1 payload, got %d", len(*received))
	}

	p := (*received)[0]
	if p.Secret != a.SecretPath {
		t.Errorf("expected secret %q, got %q", a.SecretPath, p.Secret)
	}
	if p.Level != string(LevelWarning) {
		t.Errorf("expected level %q, got %q", LevelWarning, p.Level)
	}
}

func TestWebhookNotifier_Send_NonOKStatus(t *testing.T) {
	server, _ := startFakeWebhook(t, http.StatusInternalServerError)

	n, err := NewWebhookNotifier(server.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/my-app/token",
		TimeLeft:   2 * time.Hour,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
