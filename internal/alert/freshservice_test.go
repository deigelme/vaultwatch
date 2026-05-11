package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeFreshservice(t *testing.T, statusCode int, verify func(r *http.Request, body map[string]any)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Errorf("decode body: %v", err)
		}
		if verify != nil {
			verify(r, payload)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewFreshserviceNotifier_EmptyBaseURL(t *testing.T) {
	_, err := NewFreshserviceNotifier("", "key", "user@example.com")
	if err == nil {
		t.Fatal("expected error for empty baseURL")
	}
}

func TestNewFreshserviceNotifier_EmptyAPIKey(t *testing.T) {
	_, err := NewFreshserviceNotifier("https://example.freshservice.com", "", "user@example.com")
	if err == nil {
		t.Fatal("expected error for empty apiKey")
	}
}

func TestNewFreshserviceNotifier_EmptyReporter(t *testing.T) {
	_, err := NewFreshserviceNotifier("https://example.freshservice.com", "key", "")
	if err == nil {
		t.Fatal("expected error for empty reporter")
	}
}

func TestFreshserviceNotifier_Send_Success(t *testing.T) {
	srv := startFakeFreshservice(t, http.StatusCreated, func(r *http.Request, body map[string]any) {
		if r.URL.Path != "/api/v2/tickets" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json")
		}
		ticket, ok := body["ticket"].(map[string]any)
		if !ok {
			t.Fatal("missing ticket key in payload")
		}
		if ticket["email"] != "ops@example.com" {
			t.Errorf("unexpected email %v", ticket["email"])
		}
	})
	defer srv.Close()

	n, err := NewFreshserviceNotifier(srv.URL, "testkey", "ops@example.com")
	if err != nil {
		t.Fatalf("NewFreshserviceNotifier: %v", err)
	}
	a := Alert{
		SecretPath: "secret/db/password",
		Level:      LevelWarning,
		TimeLeft:   48 * time.Hour,
		ExpireAt:   time.Now().Add(48 * time.Hour),
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestFreshserviceNotifier_Send_CriticalPriority(t *testing.T) {
	srv := startFakeFreshservice(t, http.StatusCreated, func(_ *http.Request, body map[string]any) {
		ticket := body["ticket"].(map[string]any)
		// JSON numbers decode as float64
		if ticket["priority"].(float64) != 4 {
			t.Errorf("expected priority 4 (Urgent) for critical, got %v", ticket["priority"])
		}
	})
	defer srv.Close()

	n, _ := NewFreshserviceNotifier(srv.URL, "key", "ops@example.com")
	a := Alert{Level: LevelCritical, TimeLeft: time.Hour, ExpireAt: time.Now().Add(time.Hour)}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send: %v", err)
	}
}

func TestFreshserviceNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeFreshservice(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	n, _ := NewFreshserviceNotifier(srv.URL, "badkey", "ops@example.com")
	a := Alert{Level: LevelWarning, TimeLeft: time.Hour, ExpireAt: time.Now().Add(time.Hour)}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
