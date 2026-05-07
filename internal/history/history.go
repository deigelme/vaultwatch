// Package history tracks which alerts have already been sent for a given
// secret path + level combination so the monitor does not spam notifiers
// on every poll interval.
package history

import (
	"sync"
	"time"
)

// Key uniquely identifies an alert event.
type Key struct {
	SecretPath string
	Level      string
}

// Record holds metadata about a previously fired alert.
type Record struct {
	FiredAt time.Time
}

// Tracker keeps an in-memory record of fired alerts.
type Tracker struct {
	mu      sync.RWMutex
	records map[Key]Record
	ttl     time.Duration
}

// New creates a Tracker that suppresses duplicate alerts within ttl.
func New(ttl time.Duration) *Tracker {
	return &Tracker{
		records: make(map[Key]Record),
		ttl:     ttl,
	}
}

// Seen returns true if an alert for the given key was fired within the TTL
// window and therefore should be suppressed.
func (t *Tracker) Seen(k Key) bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	r, ok := t.records[k]
	if !ok {
		return false
	}
	return time.Since(r.FiredAt) < t.ttl
}

// Record marks the key as having been alerted right now.
func (t *Tracker) Record(k Key) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.records[k] = Record{FiredAt: time.Now()}
}

// Purge removes stale entries older than the TTL to prevent unbounded growth.
func (t *Tracker) Purge() {
	t.mu.Lock()
	defer t.mu.Unlock()
	for k, r := range t.records {
		if time.Since(r.FiredAt) >= t.ttl {
			delete(t.records, k)
		}
	}
}
