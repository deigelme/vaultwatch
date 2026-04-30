// Package alert provides notification abstractions and concrete notifier
// implementations for VaultWatch secret expiration alerts.
//
// # Alert Levels
//
// Each alert carries a Level that reflects urgency:
//   - Warning: the secret is expiring within the configured warning threshold.
//   - Critical: the secret is expiring within the configured critical threshold.
//
// # Notifiers
//
// The following notifiers are available out of the box:
//
//   - StdoutNotifier   — prints alerts to standard output (useful for debugging).
//   - EmailNotifier    — sends alerts via SMTP.
//   - SlackNotifier    — posts alerts to a Slack Incoming Webhook.
//   - WebhookNotifier  — HTTP POST to a generic JSON webhook endpoint.
//   - PagerDutyNotifier — creates PagerDuty incidents via the Events API v2.
//   - OpsGenieNotifier  — creates OpsGenie alerts via the Alerts API.
//   - TeamsNotifier     — posts adaptive-card messages to Microsoft Teams.
//   - DiscordNotifier   — posts embed messages to a Discord webhook.
//   - TelegramNotifier  — sends messages via the Telegram Bot API.
//   - SNSNotifier       — publishes messages to an AWS SNS topic.
//   - VictorOpsNotifier — triggers incidents via the VictorOps (Splunk On-Call) REST endpoint.
//
// # Composing Notifiers
//
// Use NewMultiNotifier to fan-out a single alert to multiple backends:
//
//	notifier := alert.NewMultiNotifier(slackNotifier, emailNotifier, pagerdutyNotifier)
//	notifier.Send(a)
package alert
