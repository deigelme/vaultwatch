// Package alert defines the Alert type and the Notifier interface used by
// VaultWatch to dispatch secret-expiration notifications.
//
// Supported backends:
//   - StdoutNotifier: writes formatted alerts to standard output (default).
//
// Additional backends (e.g. Slack, PagerDuty, email) can be added by
// implementing the Notifier interface:
//
//	type Notifier interface {
//	    Send(alert Alert) error
//	    Name() string
//	}
//
// Alert severity levels are determined by LevelForTimeLeft, which compares the
// remaining TTL against configurable warn and critical thresholds.
package alert
