package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startFakeMatrix(t *testing.T, statusCode int, capturedBody *map[string]interface{}) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capturedBody != nil {
			_ = json.NewDecoder(r.Body).Decode(capturedBody)
		}
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(`{"event_id":"$abc123"}`))
	}))
}

func TestNewMatrixNotifier_EmptyHomeserver(t *testing.T) {
	_, err := NewMatrixNotifier("", "token", "!room:matrix.org")
	if err == nil {
		t.Fatal("expected error for empty homeserver")
	}
}

func TestNewMatrixNotifier_EmptyToken(t *testing.T) {
	_, err := NewMatrixNotifier("https://matrix.org", "", "!room:matrix.org")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestNewMatrixNotifier_EmptyRoomID(t *testing.T) {
	_, err := NewMatrixNotifier("https://matrix.org", "token", "")
	if err == nil {
		t.Fatal("expected error for empty room ID")
	}
}

func TestMatrixNotifier_Send_Success(t *testing.T) {
	var captured map[string]interface{}
	srv := startFakeMatrix(t, http.StatusOK, &captured)
	defer srv.Close()

	n, err := NewMatrixNotifier(srv.URL, "s3cr3t", "!abc:example.org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     LevelWarning,
		SecretPath: "secret/data/myapp",
		ExpiresIn: 48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if captured["msgtype"] != "m.text" {
		t.Errorf("expected msgtype m.text, got %v", captured["msgtype"])
	}
	body, _ := captured["body"].(string)
	if !strings.Contains(body, "secret/data/myapp") {
		t.Errorf("body missing secret path: %q", body)
	}
}

func TestMatrixNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeMatrix(t, http.StatusForbidden, nil)
	defer srv.Close()

	n, err := NewMatrixNotifier(srv.URL, "badtoken", "!abc:example.org")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:     LevelCritical,
		SecretPath: "secret/data/db",
		ExpiresIn: 2 * time.Hour,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
