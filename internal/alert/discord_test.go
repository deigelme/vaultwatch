package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func startFakeDiscord(t *testing.T, statusCode int, gotPayload *discordPayload) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if gotPayload != nil {
			_ = json.Unmarshal(body, gotPayload)
		}
		w.WriteHeader(statusCode)
	}))
}

func TestNewDiscordNotifier_EmptyURL(t *testing.T) {
	_, err := NewDiscordNotifier("")
	if err == nil {
		t.Fatal("expected error for empty URL")
	}
}

func TestDiscordNotifier_Send_Success(t *testing.T) {
	var got discordPayload
	srv := startFakeDiscord(t, http.StatusNoContent, &got)
	defer srv.Close()

	n, err := NewDiscordNotifier(srv.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		SecretPath: "secret/db",
		ExpireAt:   time.Now().Add(48 * time.Hour),
		TimeLeft:   48 * time.Hour,
		Level:      LevelWarning,
	}

	if err := n.Send(a); err != nil {
		t.Fatalf("Send returned error: %v", err)
	}

	if got.Username != "VaultWatch" {
		t.Errorf("expected username VaultWatch, got %q", got.Username)
	}
	if len(got.Embeds) == 0 {
		t.Fatal("expected at least one embed")
	}
	if got.Embeds[0].Color != 0xffa500 {
		t.Errorf("expected warning color 0xffa500, got 0x%x", got.Embeds[0].Color)
	}
}

func TestDiscordNotifier_Send_CriticalColor(t *testing.T) {
	var got discordPayload
	srv := startFakeDiscord(t, http.StatusNoContent, &got)
	defer srv.Close()

	n, _ := NewDiscordNotifier(srv.URL)
	a := Alert{
		SecretPath: "secret/api",
		ExpireAt:   time.Now().Add(2 * time.Hour),
		TimeLeft:   2 * time.Hour,
		Level:      LevelCritical,
	}

	_ = n.Send(a)

	if got.Embeds[0].Color != 0xff0000 {
		t.Errorf("expected critical color 0xff0000, got 0x%x", got.Embeds[0].Color)
	}
}

func TestDiscordNotifier_Send_NonOKStatus(t *testing.T) {
	srv := startFakeDiscord(t, http.StatusInternalServerError, nil)
	defer srv.Close()

	n, _ := NewDiscordNotifier(srv.URL)
	a := Alert{
		SecretPath: "secret/api",
		ExpireAt:   time.Now().Add(24 * time.Hour),
		TimeLeft:   24 * time.Hour,
		Level:      LevelWarning,
	}

	if err := n.Send(a); err == nil {
		t.Fatal("expected error for non-2xx status")
	}
}
