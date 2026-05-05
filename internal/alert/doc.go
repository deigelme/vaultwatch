// Package alert provides notification primitives for VaultWatch.
//
// # Alert Levels
//
// Alerts are categorised into three levels based on how much time remains
// before a secret expires:
//
//   - LevelInfo     – more than 7 days remaining
//   - LevelWarning  – between 1 and 7 days remaining
//   - LevelCritical – less than 24 hours remaining
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
//
//   - StdoutNotifier   – prints alerts to standard output
//   - EmailNotifier    – sends alerts via SMTP
//   - SlackNotifier    – posts alerts to a Slack webhook
//   - WebhookNotifier  – posts a JSON payload to an arbitrary HTTP endpoint
//   - PagerDutyNotifier – triggers PagerDuty incidents
//   - OpsGenieNotifier  – creates OpsGenie alerts
//   - TeamsNotifier     – sends adaptive cards to Microsoft Teams
//   - DiscordNotifier   – posts embeds to a Discord webhook
//   - TelegramNotifier  – sends messages via the Telegram Bot API
//   - SNSNotifier       – publishes to an AWS SNS topic
//   - VictorOpsNotifier – triggers VictorOps incidents
//   - SyslogNotifier    – writes alerts to the local syslog daemon
//   - MattermostNotifier – posts to a Mattermost incoming webhook
//   - GoogleChatNotifier – sends cards to a Google Chat webhook
//   - DatadogNotifier    – posts events to the Datadog Events API
//   - SplunkNotifier     – sends events to Splunk HTTP Event Collector
//   - NewRelicNotifier   – creates New Relic incidents
//   - ZendutyNotifier    – triggers Zenduty incidents
//   - AlertmanagerNotifier – posts alerts to Prometheus Alertmanager
//   - JiraNotifier       – creates Jira issues for expiring secrets
//   - MultiNotifier     – fans out to multiple notifiers simultaneously
package alert
