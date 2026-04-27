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
// Typical usage:
//
//	client, err := vault.NewClient(cfg.VaultAddress, cfg.VaultToken)
//	monitor := vault.NewMonitor(client, cfg, myAlertFunc)
//	monitor.Run(done)
package vault
