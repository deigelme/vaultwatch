// Package alert provides notifier implementations for various alerting
// backends used by VaultWatch.
//
// # Chat Ops Notifiers
//
// VaultWatch ships with first-class support for team chat platforms so that
// on-call engineers receive secret-expiration warnings directly in the channels
// they already monitor.
//
// ## Google Chat
//
// GoogleChatNotifier posts a plain-text card to a Google Chat Space via an
// incoming webhook URL.  Create a webhook under Space Settings → Apps &
// Integrations → Webhooks and supply the resulting URL:
//
//	notifier, err := alert.NewGoogleChatNotifier("https://chat.googleapis.com/v1/spaces/.../messages?key=...")
//
// ## Rocket.Chat
//
// RocketChatNotifier delivers colour-coded message attachments to any
// Rocket.Chat channel via an incoming webhook integration.  Enable the
// "Incoming WebHook" integration in your Rocket.Chat administration panel and
// use the generated webhook URL:
//
//	notifier, err := alert.NewRocketChatNotifier("https://rocketchat.example.com/hooks/...")
//
// Attachment colours follow the standard VaultWatch severity palette:
//   - Critical → red  (#FF0000)
//   - Warning  → orange (#FFA500)
//   - Info     → green  (#36A64F)
//
// Both notifiers implement the [Notifier] interface and can be composed with
// [NewMultiNotifier] to fan alerts out to multiple destinations.
package alert
