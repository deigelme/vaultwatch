// Package alert provides notification primitives for VaultWatch.
//
// It defines the Alert type, severity levels, and multiple Notifier
// implementations that can dispatch alerts through different channels:
//
//   - StdoutNotifier  – writes human-readable alerts to standard output.
//   - EmailNotifier   – sends alerts via SMTP to one or more recipients.
//
// # Alert levels
//
// Levels are derived from the time remaining before a secret expires:
//
//   - LevelCritical  – fewer than 24 hours remaining.
//   - LevelWarning   – fewer than 72 hours remaining.
//   - LevelInfo      – fewer than 168 hours (7 days) remaining.
//
// # Adding a new notifier
//
// Implement the Notifier interface:
//
//	type Notifier interface {
//		Send(a Alert) error
//	}
//
// Then wire it up in cmd/vaultwatch/main.go alongside the existing notifiers.
package alert
