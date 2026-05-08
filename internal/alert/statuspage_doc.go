// Package alert provides notifier implementations for VaultWatch alerts.
//
// # Statuspage Notifier
//
// The StatusPageNotifier integrates with Atlassian Statuspage to automatically
// update a component's status when a Vault secret is approaching expiration.
//
// ## Component Status Mapping
//
//	| Alert Level | Statuspage Status      |
//	|-------------|------------------------|
//	| info        | operational            |
//	| warning     | degraded_performance   |
//	| critical    | major_outage           |
//
// ## Configuration (vaultwatch.yaml)
//
//	alerts:
//	  - type: statuspage
//	    page_id: "abc123"
//	    component_id: "xyz789"
//	    api_key: "${STATUSPAGE_API_KEY}"
//
// ## Required Fields
//
//	- page_id:      Your Statuspage page identifier.
//	- component_id: The component to update.
//	- api_key:      OAuth API key with write access to the page.
package alert
