package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeJira(t *testing.T, statusCode int, capturedBody *jiraIssuePayload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if capturedBody != nil {
			_ = json.NewDecoder(r.Body).Decode(capturedBody)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewJiraNotifier_EmptyBaseURL(t *testing.T) {
	_, err := NewJiraNotifier("", "user", "token", "PROJ", "Task")
	if err == nil {
		t.Fatal("expected error for empty baseURL")
	}
}

func TestNewJiraNotifier_EmptyCredentials(t *testing.T) {
	_, err := NewJiraNotifier("http://jira.example.com", "", "", "PROJ", "Task")
	if err == nil {
		t.Fatal("expected error for empty credentials")
	}
}

func TestNewJiraNotifier_EmptyProjectKey(t *testing.T) {
	_, err := NewJiraNotifier("http://jira.example.com", "user", "token", "", "Task")
	if err == nil {
		t.Fatal("expected error for empty projectKey")
	}
}

func TestJiraNotifier_Send_Success(t *testing.T) {
	var captured jiraIssuePayload
	srv := startFakeJira(t, http.StatusCreated, &captured)
	defer srv.Close()

	n, err := NewJiraNotifier(srv.URL, "user", "token", "SEC", "Bug")
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
		t.Fatalf("Send() error: %v", err)
	}
	if captured.Fields.Project.Key != "SEC" {
		t.Errorf("expected project key SEC, got %s", captured.Fields.Project.Key)
	}
	if captured.Fields.IssueType.Name != "Bug" {
		t.Errorf("expected issue type Bug, got %s", captured.Fields.IssueType.Name)
	}
	if captured.Fields.Priority.Name != "Medium" {
		t.Errorf("expected priority Medium for warning, got %s", captured.Fields.Priority.Name)
	}
}

func TestJiraNotifier_Send_CriticalPriority(t *testing.T) {
	var captured jiraIssuePayload
	srv := startFakeJira(t, http.StatusCreated, &captured)
	defer srv.Close()

	n, err := NewJiraNotifier(srv.URL, "user", "token", "SEC", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/api/key",
		ExpiresAt:  time.Now().Add(6 * time.Hour),
		TimeLeft:   6 * time.Hour,
		Level:      LevelCritical,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send() error: %v", err)
	}
	if captured.Fields.Priority.Name != "Highest" {
		t.Errorf("expected priority Highest for critical, got %s", captured.Fields.Priority.Name)
	}
	if captured.Fields.IssueType.Name != "Task" {
		t.Errorf("expected default issue type Task, got %s", captured.Fields.IssueType.Name)
	}
}

func TestJiraNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeJira(t, http.StatusForbidden, nil)
	defer srv.Close()

	n, err := NewJiraNotifier(srv.URL, "user", "bad-token", "SEC", "Task")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		TimeLeft:   24 * time.Hour,
		Level:      LevelWarning,
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
