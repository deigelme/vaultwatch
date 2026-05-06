package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeGotify(t *testing.T, statusCode int, handler func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if handler != nil {
			handler(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewGotifyNotifier_EmptyBaseURL(t *testing.T) {
	_, err := NewGotifyNotifier("", "token123")
	if err == nil {
		t.Fatal("expected error for empty baseURL")
	}
}

func TestNewGotifyNotifier_EmptyToken(t *testing.T) {
	_, err := NewGotifyNotifier("http://gotify.example.com", "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestGotifyNotifier_Send_Success(t *testing.T) {
	var capturedToken string
	var capturedPayload gotifyPayload

	srv := startFakeGotify(t, http.StatusOK, func(r *http.Request) {
		capturedToken = r.URL.Query().Get("token")
		if err := json.NewDecoder(r.Body).Decode(&capturedPayload); err != nil {
			t.Errorf("decode payload: %v", err)
		}
	})
	defer srv.Close()

	n, err := NewGotifyNotifier(srv.URL, "mytoken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     LevelWarning,
		SecretPath: "secret/data/myapp",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		TimeLeft:  48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if capturedToken != "mytoken" {
		t.Errorf("expected token %q, got %q", "mytoken", capturedToken)
	}
	if capturedPayload.Priority != 5 {
		t.Errorf("expected priority 5 for warning, got %d", capturedPayload.Priority)
	}
}

func TestGotifyNotifier_Send_CriticalPriority(t *testing.T) {
	var capturedPayload gotifyPayload

	srv := startFakeGotify(t, http.StatusOK, func(r *http.Request) {
		json.NewDecoder(r.Body).Decode(&capturedPayload) //nolint:errcheck
	})
	defer srv.Close()

	n, _ := NewGotifyNotifier(srv.URL, "tok")
	a := Alert{
		Level:     LevelCritical,
		SecretPath: "secret/data/db",
		ExpiresAt: time.Now().Add(2 * time.Hour),
		TimeLeft:  2 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if capturedPayload.Priority != 10 {
		t.Errorf("expected priority 10 for critical, got %d", capturedPayload.Priority)
	}
}

func TestGotifyNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeGotify(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	n, _ := NewGotifyNotifier(srv.URL, "badtoken")
	a := Alert{
		Level:     LevelWarning,
		SecretPath: "secret/data/test",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		TimeLeft:  24 * time.Hour,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
