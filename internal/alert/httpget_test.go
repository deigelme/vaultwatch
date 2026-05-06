package alert

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func startFakeHTTPGet(t *testing.T, statusCode int) (*httptest.Server, *url.Values) {
	t.Helper()
	var received url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		received = r.URL.Query()
		w.WriteHeader(statusCode)
	}))
	t.Cleanup(srv.Close)
	return srv, &received
}

func TestNewHTTPGetNotifier_EmptyURL(t *testing.T) {
	_, err := NewHTTPGetNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestNewHTTPGetNotifier_InvalidURL(t *testing.T) {
	_, err := NewHTTPGetNotifier("not a url")
	if err == nil {
		t.Fatal("expected error for invalid URL")
	}
}

func TestHTTPGetNotifier_Send_Success(t *testing.T) {
	srv, received := startFakeHTTPGet(t, http.StatusOK)

	n, err := NewHTTPGetNotifier(srv.URL + "/notify")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db/password",
		Level:      LevelWarning,
		Message:    "expires soon",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if (*received).Get("path") != a.SecretPath {
		t.Errorf("path = %q, want %q", (*received).Get("path"), a.SecretPath)
	}
	if (*received).Get("level") != string(a.Level) {
		t.Errorf("level = %q, want %q", (*received).Get("level"), string(a.Level))
	}
	if (*received).Get("message") != a.Message {
		t.Errorf("message = %q, want %q", (*received).Get("message"), a.Message)
	}
}

func TestHTTPGetNotifier_Send_NonOKStatus(t *testing.T) {
	srv, _ := startFakeHTTPGet(t, http.StatusInternalServerError)

	n, err := NewHTTPGetNotifier(srv.URL + "/notify")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/api/key",
		Level:      LevelCritical,
		Message:    "critically close to expiry",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
