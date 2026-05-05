package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeZenduty(t *testing.T, statusCode int, capturedBody *zendutyPayload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capturedBody != nil {
			if err := json.NewDecoder(r.Body).Decode(capturedBody); err != nil {
				t.Errorf("decode body: %v", err)
			}
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewZendutyNotifier_EmptyKey(t *testing.T) {
	_, err := NewZendutyNotifier("")
	if err == nil {
		t.Fatal("expected error for empty integration key")
	}
}

func TestZendutyNotifier_Send_Success(t *testing.T) {
	var captured zendutyPayload
	srv := startFakeZenduty(t, http.StatusCreated, &captured)
	defer srv.Close()

	n, err := NewZendutyNotifier("test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.apiURL = srv.URL + "/"

	a := Alert{
		SecretPath: "secret/db/password",
		TimeLeft:   48 * time.Hour,
		Level:      Warning,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if captured.AlertType != "warning" {
		t.Errorf("expected alert_type=warning, got %q", captured.AlertType)
	}
	if captured.EntityID != "secret/db/password" {
		t.Errorf("unexpected entity_id: %q", captured.EntityID)
	}
}

func TestZendutyNotifier_Send_CriticalAlertType(t *testing.T) {
	var captured zendutyPayload
	srv := startFakeZenduty(t, http.StatusCreated, &captured)
	defer srv.Close()

	n, _ := NewZendutyNotifier("test-key")
	n.apiURL = srv.URL + "/"

	a := Alert{
		SecretPath: "secret/api/key",
		TimeLeft:   2 * time.Hour,
		Level:      Critical,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if captured.AlertType != "critical" {
		t.Errorf("expected alert_type=critical, got %q", captured.AlertType)
	}
}

func TestZendutyNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeZenduty(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, _ := NewZendutyNotifier("test-key")
	n.apiURL = srv.URL + "/"

	a := Alert{
		SecretPath: "secret/token",
		TimeLeft:   24 * time.Hour,
		Level:      Warning,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
