// Package alert — googlepubsub notifier
//
// GooglePubSubNotifier publishes VaultWatch alert events to a Google Cloud
// Pub/Sub topic. Each published message is a JSON object with the fields:
//
//	{
//	  "level":      "warning" | "critical",
//	  "secret":     "<vault secret path>",
//	  "message":    "<human-readable alert string>",
//	  "expires_at": "<RFC3339 timestamp>"
//	}
//
// # Configuration
//
// Provide a GCP project ID and Pub/Sub topic ID when constructing the notifier.
// Authentication follows Application Default Credentials (ADC) — set the
// GOOGLE_APPLICATION_CREDENTIALS environment variable or run in an environment
// with a service account attached (e.g. GKE, Cloud Run).
//
// # Example (vaultwatch.yaml)
//
//	alerts:
//	  - type: googlepubsub
//	    project_id: my-gcp-project
//	    topic_id:   vaultwatch-alerts
package alert
