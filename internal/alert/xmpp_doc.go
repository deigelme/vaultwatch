// Package alert provides Notifier implementations for various alerting
// backends. This file documents the XMPP notifier.
//
// # XMPP Notifier
//
// XMPPNotifier delivers alerts through an HTTP gateway that bridges HTTP POST
// requests to XMPP messages. This approach avoids a native XMPP client
// dependency while still supporting Jabber / XMPP infrastructure.
//
// # Configuration
//
// The notifier requires two parameters:
//
//   - gatewayURL: the HTTP endpoint of the XMPP bridge (e.g.
//     "http://xmpp-gateway.internal/send").
//   - to: the recipient Jabber ID (JID), e.g. "oncall@example.com".
//
// # Payload
//
// A JSON object is POSTed to the gateway:
//
//	{"to": "<jid>", "body": "<alert text>"}
//
// The gateway is expected to return a 2xx status code on success.
package alert
