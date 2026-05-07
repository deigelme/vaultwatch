// Package vault provides a thin wrapper around the HashiCorp Vault API client
// for use by vaultwatch.
//
// It exposes:
//
//   - Client: authenticates against a Vault server and reads secret metadata
//     (TTL, lease ID, renewability).
//
//   - SecretMeta: a lightweight struct carrying expiration information for a
//     single secret path.
//
//   - Monitor: a polling loop that checks each configured secret path on a
//     regular interval and invokes registered AlertFunc callbacks when a
//     secret's remaining TTL falls within the configured alert threshold.
//
//   - AlertFunc: a callback type invoked by Monitor when a secret's TTL
//     drops at or below the configured threshold. Implementations should
//     be non-blocking; long-running work should be dispatched to a goroutine.
//
// Error handling:
//
// Transient network errors encountered during polling are logged and skipped;
// Monitor will retry on the next interval rather than terminating. Permanent
// errors (e.g. invalid token, permission denied) are surfaced via the
// AlertFunc so callers can decide how to respond.
//
// Typical usage:
//
//	client, err := vault.NewClient(cfg.VaultAddress, cfg.VaultToken)
//	if err != nil {
//		log.Fatal(err)
//	}
//	monitor := vault.NewMonitor(client, cfg, myAlertFunc)
//	monitor.Run(done)
package vault
