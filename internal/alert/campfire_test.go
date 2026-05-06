package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startFakeCampfire(t *testing.T, statusCode int, validate func(*http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validate != nil {
			validate(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewCampfireNotifier_EmptyAccountID(t *testing.T) {
	_, err := NewCampfireNotifier("", "room1", "tok")
	if err == nil {
		t.Fatal("expected error for empty accountID")
	}
}

func TestNewCampfireNotifier_EmptyCampfireID(t *testing.T) {
	_, err := NewCampfireNotifier("acct1", "", "tok")
	if err == nil {
		t.Fatal("expected error for empty campfireID")
	}
}

func TestNewCampfireNotifier_EmptyToken(t *testing.T) {
	_, err := NewCampfireNotifier("acct1", "room1", "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestCampfireNotifier_Send_Success(t *testing.T) {
	var gotBody campfirePayload
	var gotContentType string

	srv := startFakeCampfire(t, http.StatusCreated, func(r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		body, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(body, &gotBody)
	})
	defer srv.Close()

	n, err := NewCampfireNotifier("12345", "67890", "mytoken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.baseURL = srv.URL

	a := Alert{
		Level:     LevelWarning,
		Message:   "secret/db expires in 6 days",
		Secret:    "secret/db",
		ExpiresAt: time.Now().Add(6 * 24 * time.Hour),
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if !strings.Contains(gotContentType, "application/json") {
		t.Errorf("expected JSON content-type, got %q", gotContentType)
	}
	if !strings.Contains(gotBody.Content, string(LevelWarning)) {
		t.Errorf("expected level in content, got %q", gotBody.Content)
	}
	if !strings.Contains(gotBody.Content, "secret/db expires in 6 days") {
		t.Errorf("expected message in content, got %q", gotBody.Content)
	}
}

func TestCampfireNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeCampfire(t, http.StatusForbidden, nil)
	defer srv.Close()

	n, err := NewCampfireNotifier("12345", "67890", "mytoken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.baseURL = srv.URL

	a := Alert{
		Level:   LevelCritical,
		Message: "secret/api expired",
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
