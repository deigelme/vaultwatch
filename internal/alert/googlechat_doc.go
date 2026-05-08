// Package alert provides notifier implementations for VaultWatch alerts.
//
// # Google Chat Notifier
//
// GoogleChatNotifier delivers alerts to a Google Chat space via an
// incoming webhook URL. To obtain a webhook URL, open your Google Chat
// space, click "Manage webhooks", and create a new webhook.
//
// Configuration example (vaultwatch.yaml):
//
//	alerts:
//	  - type: googlechat
//	    webhook_url: "https://chat.googleapis.com/v1/spaces/.../messages?key=...&token=..."
//
// Each alert is delivered as a plain-text card message that includes the
// alert level, secret path, time remaining, and a human-readable summary.
//
// The notifier returns an error if the webhook URL is empty or if Google
// Chat responds with a non-200 HTTP status code.
package alert
