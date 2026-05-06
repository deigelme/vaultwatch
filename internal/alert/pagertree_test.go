package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakePagerTree(t *testing.T, statusCode int, callback func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if callback != nil {
			callback(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewPagerTreeNotifier_EmptyURL(t *testing.T) {
	_, err := NewPagerTreeNotifier("")
	if err == nil {
		t.Fatal("expected error for empty integration URL, got nil")
	}
}

func TestPagerTreeNotifier_Send_Success(t *testing.T) {
	var gotBody []byte
	srv := startFakePagerTree(t, http.StatusOK, func(r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
	})
	defer srv.Close()

	n, err := NewPagerTreeNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db/password",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Level:      LevelWarning,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	var payload pagerTreePayload
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	if payload.Urgency != "medium" {
		t.Errorf("expected urgency 'medium', got %q", payload.Urgency)
	}
}

func TestPagerTreeNotifier_Send_CriticalUrgency(t *testing.T) {
	var gotBody []byte
	srv := startFakePagerTree(t, http.StatusOK, func(r *http.Request) {
		gotBody, _ = io.ReadAll(r.Body)
	})
	defer srv.Close()

	n, err := NewPagerTreeNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/api/key",
		ExpiresAt:  time.Now().Add(2 * time.Hour),
		TimeLeft:   2 * time.Hour,
		Level:      LevelCritical,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned unexpected error: %v", err)
	}

	var payload pagerTreePayload
	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("failed to parse request body: %v", err)
	}
	if payload.Urgency != "critical" {
		t.Errorf("expected urgency 'critical', got %q", payload.Urgency)
	}
}

func TestPagerTreeNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakePagerTree(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, err := NewPagerTreeNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db/password",
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Level:      LevelWarning,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status, got nil")
	}
}
