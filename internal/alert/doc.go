// Package alert provides notification primitives for VaultWatch.
//
// # Notifiers
//
// Each notifier implements the Notifier interface:
//
//	type Notifier interface {
//	    Send(a Alert) error
//	}
//
// Available notifiers:
//   - StdoutNotifier  – prints alerts to standard output
//   - EmailNotifier   – sends alerts via SMTP
//   - SlackNotifier   – posts messages to a Slack webhook
//   - WebhookNotifier – HTTP POST to a generic webhook URL
//   - PagerDutyNotifier – triggers PagerDuty incidents
//   - OpsGenieNotifier  – creates OpsGenie alerts
//   - TeamsNotifier     – posts cards to Microsoft Teams
//   - DiscordNotifier   – sends embeds to a Discord webhook
//   - TelegramNotifier  – sends messages via the Telegram Bot API
//   - SNSNotifier       – publishes to an AWS SNS topic
//   - MultiNotifier    – fans out to multiple notifiers
//
// # Alert Levels
//
// LevelForTimeLeft maps a remaining duration to a severity level:
//   - Critical : ≤ 24 h
//   - Warning  : ≤ 72 h
//   - Info     : anything longer
package alert
