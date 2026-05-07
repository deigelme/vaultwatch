// Package alert provides notification primitives for vaultwatch.
//
// # Alert
//
// An [Alert] value carries all information about an expiring secret:
// the secret path, how much time is left, the computed [Level], and
// the absolute expiry timestamp.
//
// # Levels
//
// [LevelForTimeLeft] maps a remaining duration to one of three levels:
//
//	  LevelInfo     — more than 7 days remaining
//	  LevelWarning  — between 1 and 7 days remaining
//	  LevelCritical — less than 24 hours remaining
//
// # Notifiers
//
// Every notifier implements the Notifier interface:
//
//	type Notifier interface {
//	    Send(Alert) error
//	}
//
// Available notifiers:
//
//   - [NewStdoutNotifier]    — prints to stdout (useful for debugging)
//   - [NewEmailNotifier]     — sends SMTP email
//   - [NewSlackNotifier]     — posts to a Slack incoming webhook
//   - [NewWebhookNotifier]   — HTTP POST to an arbitrary JSON endpoint
//   - [NewPagerDutyNotifier] — creates a PagerDuty event via Events API v2
//   - [NewOpsGenieNotifier]  — creates an OpsGenie alert
//   - [NewTeamsNotifier]     — posts an Adaptive Card to Microsoft Teams
//   - [NewDiscordNotifier]   — posts an embed to a Discord webhook
//   - [NewTelegramNotifier]  — sends a Telegram bot message
//   - [NewSNSNotifier]       — publishes to an AWS SNS topic
//   - [NewVictorOpsNotifier] — sends a VictorOps / Splunk On-Call alert
//   - [NewSyslogNotifier]    — writes to the local syslog daemon
//   - [NewMattermostNotifier]— posts to a Mattermost incoming webhook
//   - [NewGoogleChatNotifier]— posts to a Google Chat webhook
//   - [NewDatadogNotifier]   — creates a Datadog event
//   - [NewSplunkNotifier]    — sends an event to Splunk HEC
//   - [NewNewRelicNotifier]  — creates a New Relic alert event
//   - [NewZendutyNotifier]   — creates a Zenduty incident
//   - [NewAlertmanagerNotifier]— fires an alert to Prometheus Alertmanager
//   - [NewJiraNotifier]      — opens a Jira issue
//   - [NewServiceNowNotifier]— creates a ServiceNow incident
//   - [NewRocketChatNotifier]— posts to a Rocket.Chat webhook
//   - [NewSignalSciencesNotifier]— creates a Signal Sciences custom alert
//   - [NewGotifyNotifier]    — sends a Gotify push notification
//   - [NewMatrixNotifier]    — sends a Matrix room message
//   - [NewZulipNotifier]     — sends a Zulip message
//   - [NewPushoverNotifier]  — sends a Pushover notification
//   - [NewLarkNotifier]      — posts to a Lark (Feishu) webhook
//   - [NewPagerTreeNotifier] — creates a PagerTree alert
//   - [NewLinearNotifier]    — creates a Linear issue
//   - [NewHTTPGetNotifier]   — fires an HTTP GET request
//   - [NewGrafanaNotifier]   — creates a Grafana annotation
//   - [NewCampfireNotifier]  — posts to a Campfire room
//   - [NewNtfyNotifier]      — sends an ntfy.sh notification
//   - [NewBearyChatNotifier] — posts to a BearyChat webhook
//   - [NewSIGNL4Notifier]    — triggers a SIGNL4 alert
//
// # MultiNotifier
//
// [NewMultiNotifier] fans out a single [Alert] to multiple [Notifier]
// implementations, collecting all errors.
package alert
