// Package alert provides notifier implementations for various alerting
// backends used by vaultwatch.
//
// # File Notifier
//
// The File notifier appends a timestamped alert line to a local file on
// every Send call. It is useful for audit trails, log aggregators, or
// simple on-disk persistence of alert history.
//
// Each line written has the format:
//
//	<RFC3339 timestamp> [<level>] path=<secret-path> expires_at=<time> time_left=<duration>
//
// The file is created automatically if it does not already exist.
// Concurrent writes from multiple goroutines are safe because each Send
// call opens, writes, and closes the file independently; the OS-level
// append flag guarantees atomicity for small writes on most platforms.
//
// # Configuration
//
//	notifiers:
//	  file:
//	    path: "/var/log/vaultwatch/alerts.log"
package alert
