package vault

import (
	"context"
	"log"
	"time"

	"github.com/yourusername/vaultwatch/internal/alert"
	"github.com/yourusername/vaultwatch/internal/config"
	"github.com/yourusername/vaultwatch/internal/history"
)

// Monitor polls Vault for secret metadata and fires alerts when expiry
// thresholds are breached.
type Monitor struct {
	client   *Client
	cfg      *config.Config
	notifier alert.Notifier
	tracker  *history.Tracker
}

// NewMonitor constructs a Monitor from the given config, Vault client, and
// notifier. The history TTL is set to the poll interval so each alert fires
// at most once per cycle.
func NewMonitor(cfg *config.Config, client *Client, notifier alert.Notifier) *Monitor {
	ttl := time.Duration(cfg.Interval) * time.Second
	if ttl <= 0 {
		ttl = 60 * time.Second
	}
	return &Monitor{
		client:   client,
		cfg:      cfg,
		notifier: notifier,
		tracker:  history.New(ttl),
	}
}

// Run starts the polling loop and blocks until ctx is cancelled.
func (m *Monitor) Run(ctx context.Context) {
	ticker := time.NewTicker(time.Duration(m.cfg.Interval) * time.Second)
	defer ticker.Stop()
	m.poll()
	for {
		select {
		case <-ticker.C:
			m.poll()
		case <-ctx.Done():
			return
		}
	}
}

func (m *Monitor) poll() {
	m.tracker.Purge()
	for _, path := range m.cfg.SecretPaths {
		meta, err := m.client.GetSecretMeta(path)
		if err != nil {
			log.Printf("monitor: get secret meta %q: %v", path, err)
			continue
		}
		if meta.ExpiresAt.IsZero() {
			continue
		}
		timeLeft := time.Until(meta.ExpiresAt)
		lvl := alert.LevelForTimeLeft(timeLeft)
		if lvl == alert.LevelNone {
			continue
		}
		k := history.Key{SecretPath: path, Level: string(lvl)}
		if m.tracker.Seen(k) {
			continue
		}
		a := alert.Alert{
			SecretPath: path,
			ExpiresAt:  meta.ExpiresAt,
			TimeLeft:   timeLeft,
			Level:      lvl,
		}
		if err := m.notifier.Send(a); err != nil {
			log.Printf("monitor: send alert for %q: %v", path, err)
			continue
		}
		m.tracker.Record(k)
	}
}
