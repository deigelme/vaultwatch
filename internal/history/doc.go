// Package history provides a lightweight, thread-safe in-memory tracker for
// VaultWatch alert deduplication.
//
// # Overview
//
// When the monitor polls Vault on every interval it may encounter the same
// expiring secret many times before the secret is rotated. Without
// deduplication, every poll would fire an alert to every configured notifier,
// which quickly becomes noise.
//
// The Tracker keeps a map of (secretPath, level) → lastFiredAt timestamps.
// Before sending an alert the monitor calls Seen; if the alert was already
// sent within the configured TTL the event is skipped. After a successful
// send the monitor calls Record to update the timestamp.
//
// # TTL
//
// The TTL should be set to roughly the same value as the monitor poll
// interval (or a small multiple thereof) so that alerts are re-fired only
// after a meaningful amount of time has passed.
//
// # Purge
//
// To prevent unbounded memory growth the monitor should call Purge
// periodically (e.g. once per poll cycle) to evict entries whose TTL has
// already elapsed.
package history
