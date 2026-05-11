// Package alert provides notifier implementations for various alerting backends.
//
// # VictorOps Notifier
//
// The VictorOps notifier sends alerts to VictorOps (now Splunk On-Call) via
// its REST endpoint integration.
//
// # Configuration
//
//	[alert.victorops]
//	url = "https://alert.victorops.com/integrations/generic/20131114/alert/<routing-key>"
//
// # Message Types
//
// Alert levels are mapped to VictorOps message types:
//
//	- LevelWarning  → "WARNING"
//	- LevelCritical → "CRITICAL"
//	- (default)     → "INFO"
//
// # References
//
// https://help.victorops.com/knowledge-base/rest-endpoint-integration-guide/
package alert
