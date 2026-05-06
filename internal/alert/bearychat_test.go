package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeBearyChat(t *testing.T, statusCode int, validate func(p bearyChatPayload)) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("bearychat fake: read body: %v", err)
		}
		var p bearyChatPayload
		if err := json.Unmarshal(body, &p); err != nil {
			t.Fatalf("bearychat fake: unmarshal: %v", err)
		}
		if validate != nil {
			validate(p)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewBearyChatNotifier_EmptyURL(t *testing.T) {
	_, err := NewBearyChatNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestBearyChatNotifier_Send_Success(t *testing.T) {
	srv := startFakeBearyChat(t, http.StatusOK, func(p bearyChatPayload) {
		if p.Text == "" {
			t.Error("expected non-empty text")
		}
		if !p.Markdown {
			t.Error("expected markdown to be true")
		}
		if len(p.Attachments) != 1 {
			t.Fatalf("expected 1 attachment, got %d", len(p.Attachments))
		}
	})
	defer srv.Close()

	n, err := NewBearyChatNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	a := Alert{
		Level:      LevelWarning,
		SecretPath: "secret/my-service/token",
		TimeLeft:   48 * time.Hour,
		ExpireAt:   time.Now().Add(48 * time.Hour),
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
}

func TestBearyChatNotifier_Send_CriticalColor(t *testing.T) {
	srv := startFakeBearyChat(t, http.StatusOK, func(p bearyChatPayload) {
		if len(p.Attachments) == 0 {
			t.Fatal("expected attachments")
		}
		if p.Attachments[0].Color != "#e03e2f" {
			t.Errorf("expected critical color #e03e2f, got %s", p.Attachments[0].Color)
		}
	})
	defer srv.Close()

	n, _ := NewBearyChatNotifier(srv.URL)
	a := Alert{
		Level:      LevelCritical,
		SecretPath: "secret/db/password",
		TimeLeft:   2 * time.Hour,
		ExpireAt:   time.Now().Add(2 * time.Hour),
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBearyChatNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeBearyChat(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, _ := NewBearyChatNotifier(srv.URL)
	a := Alert{
		Level:      LevelInfo,
		SecretPath: "secret/api/key",
		TimeLeft:   120 * time.Hour,
		ExpireAt:   time.Now().Add(120 * time.Hour),
	}
	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-OK status")
	}
}
