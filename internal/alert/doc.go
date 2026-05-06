// Package alert provides notification primitives for VaultWatch.
//
// # Alert levels
//
// Each secret check produces an [Alert] whose severity is determined by
// [LevelForTimeLeft]:
//
//	- LevelInfo     – more than 7 days remaining
//	- LevelWarning  – 2–7 days remaining
//	- LevelCritical – fewer than 2 days remaining
//
// # Notifiers
//
// A Notifier is any type that implements Send(*Alert) error.  The package
// ships with the following built-in notifiers:
//
//	- [NewStdoutNotifier]       – prints to standard output (default)
//	- [NewEmailNotifier]        – SMTP email
//	- [NewSlackNotifier]        – Slack incoming webhook
//	- [NewWebhookNotifier]      – generic HTTP webhook
//	- [NewPagerDutyNotifier]    – PagerDuty Events API v2
//	- [NewOpsGenieNotifier]     – OpsGenie Alerts API
//	- [NewTeamsNotifier]        – Microsoft Teams incoming webhook
//	- [NewDiscordNotifier]      – Discord webhook
//	- [NewTelegramNotifier]     – Telegram Bot API
//	- [NewSNSNotifier]          – AWS Simple Notification Service
//	- [NewVictorOpsNotifier]    – VictorOps REST endpoint
//	- [NewSyslogNotifier]       – local or remote syslog
//	- [NewMattermostNotifier]   – Mattermost incoming webhook
//	- [NewGoogleChatNotifier]   – Google Chat webhook
//	- [NewDatadogNotifier]      – Datadog Events API
//	- [NewSplunkNotifier]       – Splunk HTTP Event Collector
//	- [NewNewRelicNotifier]     – New Relic Events API
//	- [NewZendutyNotifier]      – Zenduty alert API
//	- [NewAlertmanagerNotifier] – Prometheus Alertmanager
//	- [NewJiraNotifier]         – Jira issue creation
//	- [NewServiceNowNotifier]   – ServiceNow incident creation
//	- [NewRocketChatNotifier]   – Rocket.Chat incoming webhook
//	- [NewSignalSciencesNotifier] – Signal Sciences custom alert
//	- [NewGotifyNotifier]       – Gotify push notification
//	- [NewMatrixNotifier]       – Matrix room message
//	- [NewZulipNotifier]        – Zulip message
//	- [NewPushoverNotifier]     – Pushover push notification
//
// # Fan-out
//
// [NewMultiNotifier] wraps multiple Notifiers and delivers each alert to all
// of them, collecting any errors without short-circuiting.
package alert
