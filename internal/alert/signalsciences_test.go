package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeSignalSciences(t *testing.T, statusCode int, validate func(r *http.Request)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if validate != nil {
			validate(r)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewSignalSciencesNotifier_EmptyCorpName(t *testing.T) {
	_, err := NewSignalSciencesNotifier("", "mysite", "token123")
	if err == nil {
		t.Fatal("expected error for empty corp name")
	}
}

func TestNewSignalSciencesNotifier_EmptySiteName(t *testing.T) {
	_, err := NewSignalSciencesNotifier("mycorp", "", "token123")
	if err == nil {
		t.Fatal("expected error for empty site name")
	}
}

func TestNewSignalSciencesNotifier_EmptyToken(t *testing.T) {
	_, err := NewSignalSciencesNotifier("mycorp", "mysite", "")
	if err == nil {
		t.Fatal("expected error for empty api token")
	}
}

func TestSignalSciencesNotifier_Send_Success(t *testing.T) {
	var gotToken, gotContentType string
	var gotPayload signalSciencesPayload

	server := startFakeSignalSciences(t, http.StatusOK, func(r *http.Request) {
		gotToken = r.Header.Get("x-api-token")
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotPayload); err != nil {
			t.Errorf("failed to decode payload: %v", err)
		}
	})
	defer server.Close()

	n, err := NewSignalSciencesNotifier("mycorp", "mysite", "secret-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.endpoint = server.URL

	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/db/password",
		Expiry:     time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotToken != "secret-token" {
		t.Errorf("expected token %q, got %q", "secret-token", gotToken)
	}
	if gotContentType != "application/json" {
		t.Errorf("expected content-type application/json, got %q", gotContentType)
	}
	if gotPayload.Severity != "warning" {
		t.Errorf("expected severity 'warning', got %q", gotPayload.Severity)
	}
	if gotPayload.Source != "vaultwatch" {
		t.Errorf("expected source 'vaultwatch', got %q", gotPayload.Source)
	}
}

func TestSignalSciencesNotifier_Send_CriticalSeverity(t *testing.T) {
	var gotPayload signalSciencesPayload
	server := startFakeSignalSciences(t, http.StatusOK, func(r *http.Request) {
		json.NewDecoder(r.Body).Decode(&gotPayload) //nolint:errcheck
	})
	defer server.Close()

	n, _ := NewSignalSciencesNotifier("mycorp", "mysite", "token")
	n.endpoint = server.URL

	a := Alert{Level: LevelCritical, SecretPath: "secret/api/key", TimeLeft: 2 * time.Hour}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
	if gotPayload.Severity != "critical" {
		t.Errorf("expected severity 'critical', got %q", gotPayload.Severity)
	}
}

func TestSignalSciencesNotifier_Send_NonOKStatus(t *testing.T) {
	server := startFakeSignalSciences(t, http.StatusForbidden, nil)
	defer server.Close()

	n, _ := NewSignalSciencesNotifier("mycorp", "mysite", "token")
	n.endpoint = server.URL

	a := Alert{Level: LevelWarning, SecretPath: "secret/db", TimeLeft: 24 * time.Hour}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
