package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeStatusPage(t *testing.T, statusCode int, gotBody *statusPageBody) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gotBody != nil {
			_ = json.NewDecoder(r.Body).Decode(gotBody)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewStatusPageNotifier_EmptyPageID(t *testing.T) {
	_, err := NewStatusPageNotifier("", "comp", "key")
	if err == nil {
		t.Fatal("expected error for empty page_id")
	}
}

func TestNewStatusPageNotifier_EmptyComponentID(t *testing.T) {
	_, err := NewStatusPageNotifier("page", "", "key")
	if err == nil {
		t.Fatal("expected error for empty component_id")
	}
}

func TestNewStatusPageNotifier_EmptyAPIKey(t *testing.T) {
	_, err := NewStatusPageNotifier("page", "comp", "")
	if err == nil {
		t.Fatal("expected error for empty api_key")
	}
}

func TestStatusPageNotifier_Send_Success(t *testing.T) {
	var got statusPageBody
	srv := startFakeStatusPage(t, http.StatusOK, &got)
	defer srv.Close()

	n, err := NewStatusPageNotifier("page1", "comp1", "mykey")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.baseURL = srv.URL

	a := Alert{
		SecretPath: "secret/db",
		Level:      LevelWarning,
		ExpiresAt:  time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Message:    "expiring soon",
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if got.Component.Status != "degraded_performance" {
		t.Errorf("expected degraded_performance, got %s", got.Component.Status)
	}
}

func TestStatusPageNotifier_Send_CriticalStatus(t *testing.T) {
	var got statusPageBody
	srv := startFakeStatusPage(t, http.StatusOK, &got)
	defer srv.Close()

	n, _ := NewStatusPageNotifier("page1", "comp1", "mykey")
	n.baseURL = srv.URL

	a := Alert{Level: LevelCritical, TimeLeft: 2 * time.Hour}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if got.Component.Status != "major_outage" {
		t.Errorf("expected major_outage, got %s", got.Component.Status)
	}
}

func TestStatusPageNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeStatusPage(t, http.StatusForbidden, nil)
	defer srv.Close()

	n, _ := NewStatusPageNotifier("page1", "comp1", "mykey")
	n.baseURL = srv.URL

	a := Alert{Level: LevelWarning}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
