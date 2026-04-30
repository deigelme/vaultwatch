package alert

import (
	"fmt"
	"log/syslog"
)

// SyslogNotifier sends alerts to the local syslog daemon.
type SyslogNotifier struct {
	writer *syslog.Writer
	tag    string
}

// NewSyslogNotifier creates a SyslogNotifier that writes to syslog under the
// given tag (e.g. "vaultwatch"). Returns an error if the syslog connection
// cannot be established.
func NewSyslogNotifier(tag string) (*SyslogNotifier, error) {
	if tag == "" {
		tag = "vaultwatch"
	}
	w, err := syslog.New(syslog.LOG_DAEMON|syslog.LOG_WARNING, tag)
	if err != nil {
		return nil, fmt.Errorf("syslog: dial: %w", err)
	}
	return &SyslogNotifier{writer: w, tag: tag}, nil
}

// Send writes the alert to syslog at a priority that matches the alert level.
func (s *SyslogNotifier) Send(a Alert) error {
	msg := a.String()
	var err error
	switch a.Level {
	case LevelCritical:
		err = s.writer.Crit(msg)
	case LevelWarning:
		err = s.writer.Warning(msg)
	default:
		err = s.writer.Notice(msg)
	}
	if err != nil {
		return fmt.Errorf("syslog: write: %w", err)
	}
	return nil
}

// Close releases the underlying syslog connection.
func (s *SyslogNotifier) Close() error {
	return s.writer.Close()
}
