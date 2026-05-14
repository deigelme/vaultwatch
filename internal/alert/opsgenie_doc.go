// Package alert provides notifier implementations for VaultWatch.
//
// # OpsGenie Notifier
//
// The OpsGenieNotifier sends alert events to OpsGenie using the
// OpsGenie Alert API v2 (https://docs.opsgenie.com/docs/alert-api).
//
// # Configuration
//
// Required fields:
//   - api_key: OpsGenie API integration key.
//
// Optional fields:
//   - endpoint: override the default API URL (useful for testing or EU region).
//     Defaults to https://api.opsgenie.com/v2/alerts.
//
// # Priority Mapping
//
// VaultWatch alert levels are mapped to OpsGenie priorities as follows:
//
//	LevelCritical → P1
//	LevelWarning  → P3
//	LevelInfo     → P5
//
// # Example vaultwatch.yaml snippet
//
//	alerts:
//	  - type: opsgenie
//	    api_key: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
package alert
