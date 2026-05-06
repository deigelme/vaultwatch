package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startFakeLinear(t *testing.T, statusCode int, fn func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if fn != nil {
			fn(r)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte(`{"data":{"issueCreate":{"success":true}}}`))
	}))
}

func TestNewLinearNotifier_EmptyAPIKey(t *testing.T) {
	_, err := NewLinearNotifier("", "TEAM-1", "")
	if err == nil {
		t.Fatal("expected error for empty api key")
	}
}

func TestNewLinearNotifier_EmptyTeamID(t *testing.T) {
	_, err := NewLinearNotifier("key", "", "")
	if err == nil {
		t.Fatal("expected error for empty team id")
	}
}

func TestLinearNotifier_Send_Success(t *testing.T) {
	type gqlRequest struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}

	var captured gqlRequest
	srv := startFakeLinear(t, http.StatusOK, func(r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
		if r.Header.Get("Authorization") == "" {
			t.Error("missing Authorization header")
		}
	})
	defer srv.Close()

	n, err := NewLinearNotifier("lin_api_test", "TEAM-42", srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/payments/stripe",
		TimeLeft:   72 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	if !strings.Contains(captured.Query, "issueCreate") {
		t.Errorf("query does not contain issueCreate, got: %s", captured.Query)
	}
}

func TestLinearNotifier_Send_CriticalPriority(t *testing.T) {
	type gqlRequest struct {
		Variables map[string]interface{} `json:"variables"`
	}
	var captured gqlRequest
	srv := startFakeLinear(t, http.StatusOK, func(r *http.Request) {
		_ = json.NewDecoder(r.Body).Decode(&captured)
	})
	defer srv.Close()

	n, _ := NewLinearNotifier("key", "TEAM-1", srv.URL)
	a := Alert{Level: LevelCritical, SecretPath: "secret/crit", TimeLeft: 1 * time.Hour}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}

	input, _ := captured.Variables["input"].(map[string]interface{})
	priority, _ := input["priority"].(float64)
	if int(priority) != 1 {
		t.Errorf("priority = %v, want 1 (urgent)", priority)
	}
}

func TestLinearNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeLinear(t, http.StatusForbidden, nil)
	defer srv.Close()

	n, _ := NewLinearNotifier("bad-key", "TEAM-1", srv.URL)
	a := Alert{Level: LevelWarning, SecretPath: "secret/x", TimeLeft: 24 * time.Hour}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
