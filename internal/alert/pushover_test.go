package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakePushover(t *testing.T, statusCode int, handler func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewPushoverNotifier_EmptyAppToken(t *testing.T) {
	_, err := NewPushoverNotifier("", "user123")
	if err == nil {
		t.Fatal("expected error for empty app token")
	}
}

func TestNewPushoverNotifier_EmptyUserKey(t *testing.T) {
	_, err := NewPushoverNotifier("apptoken", "")
	if err == nil {
		t.Fatal("expected error for empty user key")
	}
}

func TestPushoverNotifier_Send_Success(t *testing.T) {
	var gotBody map[string]interface{}

	srv := startFakePushover(t, http.StatusOK, func(r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
	})
	defer srv.Close()

	n, err := NewPushoverNotifier("apptoken", "userkey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.apiURL = srv.URL

	a := Alert{
		SecretPath: "secret/db",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Level:      LevelWarning,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if gotBody["token"] != "apptoken" {
		t.Errorf("expected token 'apptoken', got %v", gotBody["token"])
	}
	if gotBody["user"] != "userkey" {
		t.Errorf("expected user 'userkey', got %v", gotBody["user"])
	}
}

func TestPushoverNotifier_Send_CriticalPriority(t *testing.T) {
	var gotBody map[string]interface{}

	srv := startFakePushover(t, http.StatusOK, func(r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
	})
	defer srv.Close()

	n, _ := NewPushoverNotifier("apptoken", "userkey")
	n.apiURL = srv.URL

	a := Alert{
		SecretPath: "secret/db",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
		TimeLeft:   2 * time.Hour,
		Level:      LevelCritical,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	priority, ok := gotBody["priority"].(float64)
	if !ok || int(priority) != 1 {
		t.Errorf("expected priority 1 for critical, got %v", gotBody["priority"])
	}
}

func TestPushoverNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakePushover(t, http.StatusBadRequest, nil)
	defer srv.Close()

	n, _ := NewPushoverNotifier("apptoken", "userkey")
	n.apiURL = srv.URL

	a := Alert{
		SecretPath: "secret/db",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		TimeLeft:   24 * time.Hour,
		Level:      LevelWarning,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
