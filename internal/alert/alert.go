// Package alert provides notification backends for VaultWatch secret expiration alerts.
package alert

import (
	"fmt"
	"time"
)

// Level represents the severity of an alert.
type Level string

const (
	LevelInfo    Level = "info"
	LevelWarning Level = "warning"
	LevelCritical Level = "critical"
)

// Alert holds the data for a single secret expiration notification.
type Alert struct {
	SecretPath string
	ExpiresAt  time.Time
	TimeLeft   time.Duration
	Level      Level
}

// String returns a human-readable representation of the alert.
func (a Alert) String() string {
	return fmt.Sprintf("[%s] Secret '%s' expires in %s (at %s)",
		a.Level,
		a.SecretPath,
		a.TimeLeft.Round(time.Second),
		a.ExpiresAt.Format(time.RFC3339),
	)
}

// Notifier is the interface implemented by all alert backends.
type Notifier interface {
	Send(alert Alert) error
	Name() string
}

// LevelForTimeLeft returns an alert level based on how much time remains.
func LevelForTimeLeft(d time.Duration, warnThreshold, criticalThreshold time.Duration) Level {
	switch {
	case d <= criticalThreshold:
		return LevelCritical
	case d <= warnThreshold:
		return LevelWarning
	default:
		return LevelInfo
	}
}
