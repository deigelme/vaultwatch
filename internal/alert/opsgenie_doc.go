// Package alert provides notifier implementations for various alerting
// backends used by vaultwatch.
//
// # OpsGenie Notifier
//
// The OpsGenie notifier sends alerts to the OpsGenie Alerts API.
// It maps vaultwatch alert levels to OpsGenie priority values:
//
//   - Critical → P1
//   - Warning  → P3
//   - Info     → P5
//
// # Configuration
//
// Required fields:
//
//	notifiers:
//	  opsgenie:
//	    api_key: "<your-opsgenie-api-key>"
//
// Optional fields:
//
//	notifiers:
//	  opsgenie:
//	    api_key: "<your-opsgenie-api-key>"
//	    endpoint: "https://api.eu.opsgenie.com/v2/alerts"  # default: api.opsgenie.com
//
// The notifier sets the alert alias to the secret path so that OpsGenie
// can de-duplicate repeated alerts for the same secret.
package alert
