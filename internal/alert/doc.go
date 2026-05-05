// Package alert provides notification types and notifier implementations
// for VaultWatch secret expiration alerts.
//
// # Alert Levels
//
// Three severity levels are defined:
//   - [LevelInfo]     — expiration is far away (informational)
//   - [LevelWarning]  — expiration is approaching
//   - [LevelCritical] — expiration is imminent
//
// Use [LevelForTimeLeft] to derive the appropriate level from a duration.
//
// # Notifiers
//
// Each notifier implements the Notifier interface:
//
//	type Notifier interface {
//	    Send(Alert) error
//	}
//
// Available notifiers:
//   - [NewStdoutNotifier]    — writes alerts to standard output
//   - [NewEmailNotifier]     — sends alerts via SMTP
//   - [NewSlackNotifier]     — posts to a Slack incoming webhook
//   - [NewWebhookNotifier]   — HTTP POST to an arbitrary webhook
//   - [NewPagerDutyNotifier] — creates PagerDuty incidents
//   - [NewOpsGenieNotifier]  — creates OpsGenie alerts
//   - [NewTeamsNotifier]     — posts to Microsoft Teams
//   - [NewDiscordNotifier]   — posts to a Discord webhook
//   - [NewTelegramNotifier]  — sends via Telegram Bot API
//   - [NewSNSNotifier]       — publishes to AWS SNS
//   - [NewVictorOpsNotifier] — triggers VictorOps incidents
//   - [NewSyslogNotifier]    — writes to the system syslog
//   - [NewMattermostNotifier]— posts to a Mattermost webhook
//   - [NewGoogleChatNotifier]— posts to a Google Chat webhook
//
// Use [NewMultiNotifier] to fan out a single alert to multiple notifiers.
package alert
