package alert

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startFakeZulip(t *testing.T, statusCode int, capturedBody *string) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capturedBody != nil {
			b, _ := io.ReadAll(r.Body)
			*capturedBody = string(b)
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(`{"result":"success","id":1}`))
	}))
}

func TestNewZulipNotifier_EmptyBaseURL(t *testing.T) {
	_, err := NewZulipNotifier("", "bot@example.com", "key", "alerts", "vault")
	if err == nil {
		t.Fatal("expected error for empty base URL")
	}
}

func TestNewZulipNotifier_EmptyBot(t *testing.T) {
	_, err := NewZulipNotifier("https://z.example.com", "", "key", "alerts", "vault")
	if err == nil {
		t.Fatal("expected error for empty bot email")
	}
}

func TestNewZulipNotifier_EmptyAPIKey(t *testing.T) {
	_, err := NewZulipNotifier("https://z.example.com", "bot@example.com", "", "alerts", "vault")
	if err == nil {
		t.Fatal("expected error for empty API key")
	}
}

func TestNewZulipNotifier_EmptyStream(t *testing.T) {
	_, err := NewZulipNotifier("https://z.example.com", "bot@example.com", "key", "", "vault")
	if err == nil {
		t.Fatal("expected error for empty stream")
	}
}

func TestZulipNotifier_Send_Success(t *testing.T) {
	var captured string
	srv := startFakeZulip(t, http.StatusOK, &captured)
	defer srv.Close()

	n, err := NewZulipNotifier(srv.URL, "bot@example.com", "apikey", "vault-alerts", "expiry")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/data/app",
		ExpiresIn:  72 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if !strings.Contains(captured, "secret%2Fdata%2Fapp") && !strings.Contains(captured, "secret/data/app") {
		t.Errorf("captured body missing secret path: %q", captured)
	}
	if !strings.Contains(captured, "stream") {
		t.Errorf("captured body missing type=stream: %q", captured)
	}
}

func TestZulipNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeZulip(t, http.StatusUnauthorized, nil)
	defer srv.Close()

	n, err := NewZulipNotifier(srv.URL, "bot@example.com", "badkey", "vault-alerts", "expiry")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/data/db",
		ExpiresIn:  1 * time.Hour,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
