package alert

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func startFakeTelegram(t *testing.T, ok bool) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(telegramResponse{OK: ok})
	}))
}

func TestNewTelegramNotifier_EmptyToken(t *testing.T) {
	_, err := NewTelegramNotifier("", "123456")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestNewTelegramNotifier_EmptyChatID(t *testing.T) {
	_, err := NewTelegramNotifier("bot-token", "")
	if err == nil {
		t.Fatal("expected error for empty chat ID")
	}
}

func TestTelegramNotifier_Send_Success(t *testing.T) {
	srv := startFakeTelegram(t, true)
	defer srv.Close()

	n, err := NewTelegramNotifier("testtoken", "-100123456")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	n.apiBase = srv.URL // override to hit fake server

	a := Alert{
		SecretPath: "secret/db",
		ExpireAt:   time.Now().Add(72 * time.Hour),
		TimeLeft:   72 * time.Hour,
		Level:      LevelInfo,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}
}

func TestTelegramNotifier_Send_APIError(t *testing.T) {
	srv := startFakeTelegram(t, false)
	defer srv.Close()

	n, _ := NewTelegramNotifier("testtoken", "-100123456")
	n.apiBase = srv.URL

	a := Alert{
		SecretPath: "secret/api",
		ExpireAt:   time.Now().Add(1 * time.Hour),
		TimeLeft:   1 * time.Hour,
		Level:      LevelCritical,
	}

	err := n.Send(a)
	if err == nil {
		t.Fatal("expected error when API returns ok=false")
	}
	if !strings.Contains(err.Error(), "ok=false") {
		t.Errorf("unexpected error message: %v", err)
	}
}
