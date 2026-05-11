// Package alert provides notifier implementations for various alerting backends.
//
// # Signal Sciences (Fastly Next-Gen WAF) Notifier
//
// The Signal Sciences notifier creates custom events in the Signal Sciences
// platform via its API, useful for correlating Vault secret expiration events
// with WAF activity.
//
// # Configuration
//
//	[alert.signalsciences]
//	corp_name = "my-corp"
//	site_name = "my-site"
//	token     = "<api-token>"
//
// # Severity Mapping
//
// Alert levels are mapped to Signal Sciences event types:
//
//	- LevelWarning  → "warning"
//	- LevelCritical → "error"
//	- (default)     → "info"
//
// # References
//
// https://docs.fastly.com/signalsciences/developer/using-the-api/
package alert
