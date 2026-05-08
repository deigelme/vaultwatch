package alert

import (
	"strings"
	"testing"
	"time"
)

func TestNewExecNotifier_EmptyCommand(t *testing.T) {
	_, err := NewExecNotifier("", nil, 5*time.Second)
	if err == nil {
		t.Fatal("expected error for empty command, got nil")
	}
}

func TestNewExecNotifier_DefaultTimeout(t *testing.T) {
	n, err := NewExecNotifier("echo", nil, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n.timeout != 30*time.Second {
		t.Errorf("expected default timeout 30s, got %v", n.timeout)
	}
}

func TestExecNotifier_Send_Success(t *testing.T) {
	n, err := NewExecNotifier("echo", []string{"-n"}, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Path:     "secret/my-app/db",
		Level:    LevelWarning,
		TimeLeft: 48 * 3600,
	}
	if err := n.Send(a); err != nil {
		t.Errorf("unexpected error from Send: %v", err)
	}
}

func TestExecNotifier_Send_CommandNotFound(t *testing.T) {
	n, err := NewExecNotifier("/nonexistent/binary-xyz", nil, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Path:     "secret/my-app/db",
		Level:    LevelCritical,
		TimeLeft: 3600,
	}
	err = n.Send(a)
	if err == nil {
		t.Fatal("expected error for missing binary, got nil")
	}
	if !strings.Contains(err.Error(), "exec notifier") {
		t.Errorf("error message should contain 'exec notifier', got: %v", err)
	}
}

func TestExecNotifier_Send_NonZeroExit(t *testing.T) {
	n, err := NewExecNotifier("false", nil, 5*time.Second)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Path:     "secret/my-app/db",
		Level:    LevelWarning,
		TimeLeft: 72 * 3600,
	}
	err = n.Send(a)
	if err == nil {
		t.Fatal("expected error for non-zero exit, got nil")
	}
	if !strings.Contains(err.Error(), "exec notifier") {
		t.Errorf("error message should contain 'exec notifier', got: %v", err)
	}
}

func TestExecNotifier_Send_Timeout(t *testing.T) {
	n, err := NewExecNotifier("sleep", []string{"10"}, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	a := Alert{
		Path:     "secret/my-app/db",
		Level:    LevelCritical,
		TimeLeft: 1800,
	}
	err = n.Send(a)
	if err == nil {
		t.Fatal("expected timeout error, got nil")
	}
}
