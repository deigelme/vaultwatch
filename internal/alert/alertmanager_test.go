package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeAlertmanager(t *testing.T, statusCode int, capturedBody *[]amAlert) *httptest.Server {
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

func TestNewAlertmanagerNotifier_EmptyEndpoint(t *testing.T) {
	_, err := NewAlertmanagerNotifier("")
	if err == nil {
		t.Fatal("expected error for empty endpoint")
	}
}

func TestAlertmanagerNotifier_Send_Success(t *testing.T) {
	var captured []amAlert
	srv := startFakeAlertmanager(t, http.StatusOK, &captured)
	defer srv.Close()

	n, err := NewAlertmanagerNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db/password",
		TimeLeft:   48 * time.Hour,
		Level:      Warning,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if len(captured) != 1 {
		t.Fatalf("expected 1 alert, got %d", len(captured))
	}
	if captured[0].Labels["severity"] != "warning" {
		t.Errorf("expected severity=warning, got %q", captured[0].Labels["severity"])
	}
	if captured[0].Labels["secret"] != "secret/db/password" {
		t.Errorf("unexpected secret label: %q", captured[0].Labels["secret"])
	}
}

func TestAlertmanagerNotifier_Send_CriticalSeverity(t *testing.T) {
	var captured []amAlert
	srv := startFakeAlertmanager(t, http.StatusOK, &captured)
	defer srv.Close()

	n, _ := NewAlertmanagerNotifier(srv.URL)

	a := Alert{
		SecretPath: "secret/api/key",
		TimeLeft:   1 * time.Hour,
		Level:      Critical,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if captured[0].Labels["severity"] != "critical" {
		t.Errorf("expected severity=critical, got %q", captured[0].Labels["severity"])
	}
}

func TestAlertmanagerNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeAlertmanager(t, http.StatusBadGateway, nil)
	defer srv.Close()

	n, _ := NewAlertmanagerNotifier(srv.URL)

	a := Alert{
		SecretPath: "secret/token",
		TimeLeft:   24 * time.Hour,
		Level:      Warning,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
