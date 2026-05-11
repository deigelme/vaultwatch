// Package alert provides notifier implementations for vaultwatch alerts.
//
// # SignalWire Notifier
//
// The SignalWireNotifier delivers vault secret expiration alerts as SMS
// messages using the SignalWire messaging REST API (LaML-compatible).
//
// # Configuration
//
// Required fields:
//
//	space_url   – Your SignalWire space URL, e.g. https://myspace.signalwire.com
//	project_id  – The SignalWire project (account) SID
//	api_token   – The REST API token for the project
//	from        – The E.164 source phone number (must be a SignalWire number)
//	to          – The E.164 destination phone number
//
// # Example vaultwatch.yaml snippet
//
//	alerts:
//	  - type: signalwire
//	    space_url: "https://myspace.signalwire.com"
//	    project_id: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
//	    api_token: "PTxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
//	    from: "+15550001234"
//	    to: "+15559876543"
package alert
