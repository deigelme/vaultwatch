package vault

import (
	"log"
	"time"

	"github.com/yourusername/vaultwatch/internal/config"
)

// AlertFunc is called when a secret is approaching expiration.
type AlertFunc func(meta *SecretMeta, warningThreshold time.Duration)

// Monitor polls Vault secret paths on a configured interval.
type Monitor struct {
	client   *Client
	cfg      *config.Config
	alertFns []AlertFunc
}

// NewMonitor creates a Monitor wired to the given client and config.
func NewMonitor(client *Client, cfg *config.Config, alertFns ...AlertFunc) *Monitor {
	return &Monitor{
		client:   client,
		cfg:      cfg,
		alertFns: alertFns,
	}
}

// Run starts the polling loop and blocks until the done channel is closed.
func (m *Monitor) Run(done <-chan struct{}) {
	ticker := time.NewTicker(m.cfg.CheckInterval)
	defer ticker.Stop()

	log.Printf("vaultwatch: starting monitor, interval=%s", m.cfg.CheckInterval)

	for {
		select {
		case <-ticker.C:
			m.checkAll()
		case <-done:
			log.Println("vaultwatch: monitor stopped")
			return
		}
	}
}

// checkAll iterates all configured secret paths and fires alerts as needed.
func (m *Monitor) checkAll() {
	threshold := m.cfg.AlertThreshold

	for _, path := range m.cfg.SecretPaths {
		meta, err := m.client.GetSecretMeta(path)
		if err != nil {
			log.Printf("vaultwatch: error reading %q: %v", path, err)
			continue
		}

		if time.Until(meta.ExpiresAt) <= threshold {
			for _, fn := range m.alertFns {
				fn(meta, threshold)
			}
		}
	}
}
