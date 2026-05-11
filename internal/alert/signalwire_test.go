package alert

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"
)

func startFakeSignalWire(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}))
}

func TestNewSignalWireNotifier_EmptySpaceURL(t *testing.T) {
	_, err := NewSignalWireNotifier("", "proj", "token", "+10000000000", "+19999999999")
	if err == nil {
		t.Fatal("expected error for empty spaceURL")
	}
}

func TestNewSignalWireNotifier_EmptyProjectID(t *testing.T) {
	_, err := NewSignalWireNotifier("https://x.signalwire.com", "", "token", "+10000000000", "+19999999999")
	if err == nil {
		t.Fatal("expected error for empty projectID")
	}
}

func TestNewSignalWireNotifier_EmptyAPIToken(t *testing.T) {
	_, err := NewSignalWireNotifier("https://x.signalwire.com", "proj", "", "+10000000000", "+19999999999")
	if err == nil {
		t.Fatal("expected error for empty apiToken")
	}
}

func TestNewSignalWireNotifier_EmptyFrom(t *testing.T) {
	_, err := NewSignalWireNotifier("https://x.signalwire.com", "proj", "token", "", "+19999999999")
	if err == nil {
		t.Fatal("expected error for empty from")
	}
}

func TestNewSignalWireNotifier_EmptyTo(t *testing.T) {
	_, err := NewSignalWireNotifier("https://x.signalwire.com", "proj", "token", "+10000000000", "")
	if err == nil {
		t.Fatal("expected error for empty to")
	}
}

func TestSignalWireNotifier_Send_Success(t *testing.T) {
	srv := startFakeSignalWire(t, http.StatusCreated)
	defer srv.Close()

	// Rewrite the notifier's spaceURL to point at the test server.
	// We also override the internal HTTP client so it doesn't follow
	// the real SignalWire domain.
	parsed, _ := url.Parse(srv.URL)
	_ = parsed

	n, err := NewSignalWireNotifier(srv.URL, "proj123", "tok", "+10000000001", "+10000000002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     LevelWarning,
		Path:      "secret/myapp/db",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		TimeLeft:  48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() returned error: %v", err)
	}
}

func TestSignalWireNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeSignalWire(t, http.StatusForbidden)
	defer srv.Close()

	n, err := NewSignalWireNotifier(srv.URL, "proj123", "tok", "+10000000001", "+10000000002")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     LevelCritical,
		Path:      "secret/myapp/db",
		ExpiresAt: time.Now().Add(2 * time.Hour),
		TimeLeft:  2 * time.Hour,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
