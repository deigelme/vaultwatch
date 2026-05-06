package alert_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

func startFakeServiceNow(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv
}

func TestNewServiceNowNotifier_EmptyBaseURL(t *testing.T) {
	_, err := alert.NewServiceNowNotifier("", "user", "pass", "incident")
	if err == nil {
		t.Fatal("expected error for empty base URL")
	}
}

func TestNewServiceNowNotifier_EmptyCredentials(t *testing.T) {
	_, err := alert.NewServiceNowNotifier("https://example.service-now.com", "", "", "incident")
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}
}

func TestServiceNowNotifier_Send_Success(t *testing.T) {
	var received map[string]interface{}

	srv := startFakeServiceNow(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"result": map[string]string{"sys_id": "abc123"}})
	})

	n, err := alert.NewServiceNowNotifier(srv.URL, "admin", "password", "incident")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := alert.Alert{
		Level:     alert.Warning,
		Path:      "secret/myapp/db",
		ExpiresAt: time.Now().Add(72 * time.Hour),
		TimeLeft:  72 * time.Hour,
		Message:   "Secret expiring in 72h",
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if received["short_description"] == nil {
		t.Error("expected short_description in payload")
	}
}

func TestServiceNowNotifier_Send_CriticalUrgency(t *testing.T) {
	var received map[string]interface{}

	srv := startFakeServiceNow(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"result": map[string]string{"sys_id": "xyz"}})
	})

	n, _ := alert.NewServiceNowNotifier(srv.URL, "admin", "pass", "incident")

	a := alert.Alert{
		Level:     alert.Critical,
		Path:      "secret/prod/key",
		ExpiresAt: time.Now().Add(6 * time.Hour),
		TimeLeft:  6 * time.Hour,
		Message:   "Secret expiring in 6h",
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	urgency, ok := received["urgency"].(string)
	if !ok || urgency != "1" {
		t.Errorf("expected urgency=1 for critical, got %v", received["urgency"])
	}
}

func TestServiceNowNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeServiceNow(t, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "forbidden", http.StatusForbidden)
	})

	n, _ := alert.NewServiceNowNotifier(srv.URL, "admin", "pass", "incident")

	a := alert.Alert{
		Level:   alert.Warning,
		Path:    "secret/app/token",
		Message: "expiring soon",
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error on non-OK HTTP status")
	}
}
