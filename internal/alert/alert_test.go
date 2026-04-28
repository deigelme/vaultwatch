package alert_test

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
)

func TestLevelForTimeLeft(t *testing.T) {
	warn := 24 * time.Hour
	crit := 6 * time.Hour

	cases := []struct {
		name     string
		left     time.Duration
		expected alert.Level
	}{
		{"info", 48 * time.Hour, alert.LevelInfo},
		{"warning", 12 * time.Hour, alert.LevelWarning},
		{"critical", 2 * time.Hour, alert.LevelCritical},
		{"exactly warn boundary", 24 * time.Hour, alert.LevelWarning},
		{"exactly crit boundary", 6 * time.Hour, alert.LevelCritical},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := alert.LevelForTimeLeft(tc.left, warn, crit)
			if got != tc.expected {
				t.Errorf("LevelForTimeLeft(%v) = %v, want %v", tc.left, got, tc.expected)
			}
		})
	}
}

func TestAlertString(t *testing.T) {
	now := time.Now().Add(3 * time.Hour)
	a := alert.Alert{
		SecretPath: "secret/db/password",
		ExpiresAt:  now,
		TimeLeft:   3 * time.Hour,
		Level:      alert.LevelCritical,
	}
	s := a.String()
	if !strings.Contains(s, "critical") {
		t.Errorf("expected 'critical' in alert string, got: %s", s)
	}
	if !strings.Contains(s, "secret/db/password") {
		t.Errorf("expected secret path in alert string, got: %s", s)
	}
}

func TestStdoutNotifier_Send(t *testing.T) {
	var buf bytes.Buffer
	n := &alert.StdoutNotifier{Writer: &buf}

	if n.Name() != "stdout" {
		t.Errorf("expected name 'stdout', got %s", n.Name())
	}

	a := alert.Alert{
		SecretPath: "secret/api/key",
		ExpiresAt:  time.Now().Add(1 * time.Hour),
		TimeLeft:   1 * time.Hour,
		Level:      alert.LevelWarning,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(buf.String(), "secret/api/key") {
		t.Errorf("expected secret path in output, got: %s", buf.String())
	}
}
