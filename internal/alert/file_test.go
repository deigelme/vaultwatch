package alert

import (
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewFileNotifier_EmptyPath(t *testing.T) {
	_, err := NewFileNotifier("")
	if err == nil {
		t.Fatal("expected error for empty path, got nil")
	}
}

func TestFileNotifier_Send_Success(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.log")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmp.Close()

	n, err := NewFileNotifier(tmp.Name())
	if err != nil {
		t.Fatalf("NewFileNotifier: %v", err)
	}

	a := Alert{
		Level:     LevelWarning,
		Path:      "secret/myapp/db",
		ExpiresAt: time.Now().Add(48 * time.Hour),
		TimeLeft:  48 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send: %v", err)
	}

	data, err := os.ReadFile(tmp.Name())
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	contents := string(data)
	if !strings.Contains(contents, "secret/myapp/db") {
		t.Errorf("expected path in output, got: %s", contents)
	}
	if !strings.Contains(contents, string(LevelWarning)) {
		t.Errorf("expected level in output, got: %s", contents)
	}
}

func TestFileNotifier_Send_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/new-alerts.log"

	n, err := NewFileNotifier(path)
	if err != nil {
		t.Fatalf("NewFileNotifier: %v", err)
	}

	a := Alert{
		Level:     LevelCritical,
		Path:      "secret/creds",
		ExpiresAt: time.Now().Add(2 * time.Hour),
		TimeLeft:  2 * time.Hour,
	}
	if err := n.Send(a); err != nil {
		t.Fatalf("Send: %v", err)
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("expected file to be created")
	}
}

func TestFileNotifier_Send_Appends(t *testing.T) {
	tmp, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.log")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	tmp.Close()

	n, err := NewFileNotifier(tmp.Name())
	if err != nil {
		t.Fatalf("NewFileNotifier: %v", err)
	}

	for i := 0; i < 3; i++ {
		a := Alert{
			Level:     LevelInfo,
			Path:      "secret/token",
			ExpiresAt: time.Now().Add(72 * time.Hour),
			TimeLeft:  72 * time.Hour,
		}
		if err := n.Send(a); err != nil {
			t.Fatalf("Send #%d: %v", i, err)
		}
	}

	data, _ := os.ReadFile(tmp.Name())
	lines := strings.Count(string(data), "\n")
	if lines != 3 {
		t.Errorf("expected 3 lines, got %d", lines)
	}
}
