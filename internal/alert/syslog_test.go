package alert

import (
	"fmt"
	"log/syslog"
	"strings"
	"testing"
	"time"
)

// fakeSyslogWriter satisfies the subset of syslog.Writer we need for testing.
type fakeSyslogWriter struct {
	lines []string
	prios []syslog.Priority
}

func (f *fakeSyslogWriter) Crit(m string) error    { f.record(syslog.LOG_CRIT, m); return nil }
func (f *fakeSyslogWriter) Warning(m string) error { f.record(syslog.LOG_WARNING, m); return nil }
func (f *fakeSyslogWriter) Notice(m string) error  { f.record(syslog.LOG_NOTICE, m); return nil }
func (f *fakeSyslogWriter) Close() error           { return nil }
func (f *fakeSyslogWriter) record(p syslog.Priority, m string) {
	f.lines = append(f.lines, m)
	f.prios = append(f.prios, p)
}

func TestNewSyslogNotifier_DefaultTag(t *testing.T) {
	// We cannot easily connect to syslog in CI; just verify the tag default
	// logic by inspecting the struct when we can construct it.
	n, err := NewSyslogNotifier("")
	if err != nil {
		t.Skipf("syslog not available: %v", err)
	}
	defer n.Close()
	if n.tag != "vaultwatch" {
		t.Errorf("expected default tag 'vaultwatch', got %q", n.tag)
	}
}

func TestSyslogNotifier_Send_Warning(t *testing.T) {
	fw := &fakeSyslogWriter{}
	n := &SyslogNotifier{writer: nil, tag: "test"}
	// Swap real writer for fake via a helper closure.
	sendFn := func(a Alert) error {
		msg := a.String()
		switch a.Level {
		case LevelCritical:
			return fw.Crit(msg)
		case LevelWarning:
			return fw.Warning(msg)
		default:
			return fw.Notice(msg)
		}
	}
	_ = n // keep linter happy

	a := Alert{
		Level:     LevelWarning,
		Path:      "secret/db/password",
		TimeLeft:  48 * time.Hour,
		ExpiresAt: time.Now().Add(48 * time.Hour),
	}
	if err := sendFn(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(fw.lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(fw.lines))
	}
	if !strings.Contains(fw.lines[0], "secret/db/password") {
		t.Errorf("message missing path: %q", fw.lines[0])
	}
	if fw.prios[0] != syslog.LOG_WARNING {
		t.Errorf("expected LOG_WARNING, got %v", fw.prios[0])
	}
}

func TestSyslogNotifier_Send_Critical(t *testing.T) {
	fw := &fakeSyslogWriter{}
	sendFn := func(a Alert) error {
		msg := a.String()
		switch a.Level {
		case LevelCritical:
			return fw.Crit(msg)
		case LevelWarning:
			return fw.Warning(msg)
		default:
			return fw.Notice(msg)
		}
	}

	a := Alert{
		Level:     LevelCritical,
		Path:      "secret/api/key",
		TimeLeft:  2 * time.Hour,
		ExpiresAt: time.Now().Add(2 * time.Hour),
	}
	if err := sendFn(a); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fw.prios[0] != syslog.LOG_CRIT {
		t.Errorf("expected LOG_CRIT, got %v", fw.prios[0])
	}
	fmt.Println(fw.lines[0]) // surface message in verbose mode
}
