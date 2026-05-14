// Package alert provides notifier implementations for VaultWatch.
//
// # Google Chat Notifier
//
// GoogleChatNotifier delivers secret-expiration alerts to a Google Chat space
// via an incoming webhook URL.
//
// # Configuration
//
// Obtain a webhook URL from the Google Chat space settings under
// "Apps & Integrations" → "Manage webhooks".
//
// Example YAML:
//
//	notifiers:
//	  googlechat:
//	    webhook_url: "https://chat.googleapis.com/v1/spaces/.../messages?key=...&token=..."
//
// # Alert format
//
// Messages are sent as plain text with the alert level and secret path.
package alert
