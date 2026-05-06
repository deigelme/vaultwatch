package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakePagerDuty(t *testing.T, statusCode int, fn func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewPagerDutyNotifier_EmptyKey(t *testing.T) {
	_, err := NewPagerDutyNotifier("")
	if err == nil {
		t.Fatal("expected error for empty integration key")
	}
}

func TestPagerDutyNotifier_Send_Success(t *testing.T) {
	var gotPayload pagerDutyPayload

	srv := startFakePagerDuty(t, http.StatusAccepted, func(r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotPayload)
	})
	defer srv.Close()

	n, err := NewPagerDutyNotifier("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.endpoint = srv.URL

	a := Alert{
		Path:      "secret/myapp/db",
		Level:     LevelWarning,
		ExpiresAt: time.Now().Add(48 * time.Hour),
		TimeLeft:  48 * time.Hour,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if gotPayload.RoutingKey != "test-key" {
		t.Errorf("routing_key = %q, want %q", gotPayload.RoutingKey, "test-key")
	}
	if gotPayload.EventAction != "trigger" {
		t.Errorf("event_action = %q, want trigger", gotPayload.EventAction)
	}
	if gotPayload.Payload.Severity != "warning" {
		t.Errorf("severity = %q, want warning", gotPayload.Payload.Severity)
	}
	if gotPayload.Payload.Source != "vaultwatch" {
		t.Errorf("source = %q, want vaultwatch", gotPayload.Payload.Source)
	}
}

func TestPagerDutyNotifier_Send_CriticalSeverity(t *testing.T) {
	srv := startFakePagerDuty(t, http.StatusAccepted, nil)
	defer srv.Close()

	n, _ := NewPagerDutyNotifier("key")
	n.endpoint = srv.URL

	a := Alert{Level: LevelCritical, ExpiresAt: time.Now().Add(2 * time.Hour), TimeLeft: 2 * time.Hour}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
}

func TestPagerDutyNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakePagerDuty(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, _ := NewPagerDutyNotifier("key")
	n.endpoint = srv.URL

	a := Alert{Level: LevelWarning, ExpiresAt: time.Now().Add(24 * time.Hour), TimeLeft: 24 * time.Hour}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
