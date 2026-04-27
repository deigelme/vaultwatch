package vault_test

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/config"
	"github.com/yourusername/vaultwatch/internal/vault"
)

type stubClient struct {
	meta *vault.SecretMeta
	err  error
}

// We test Monitor behaviour by embedding a fake via interface in a real scenario;
// here we verify the alertFn is triggered when TTL is within threshold.
func TestMonitor_AlertFiredWhenExpiringSoon(t *testing.T) {
	var alertCount int64

	cfg := &config.Config{
		VaultAddress:   "http://127.0.0.1:8200",
		SecretPaths:    []string{"secret/data/myapp"},
		CheckInterval:  50 * time.Millisecond,
		AlertThreshold: 10 * time.Minute,
	}

	// Build a mock server that returns a soon-expiring secret (TTL = 5 min).
	server := newMockVaultServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"lease_id": "abc",
			"renewable": false,
			"lease_duration": 300,
			"data": {"key": "val"}
		}`))
	})
	defer server.Close()

	cfg.VaultAddress = server.URL

	client, err := vault.NewClient(server.URL, "token")
	if err != nil {
		t.Fatalf("client error: %v", err)
	}

	alertFn := func(meta *vault.SecretMeta, threshold time.Duration) {
		atomic.AddInt64(&alertCount, 1)
	}

	monitor := vault.NewMonitor(client, cfg, alertFn)

	done := make(chan struct{})
	go monitor.Run(done)

	time.Sleep(200 * time.Millisecond)
	close(done)
	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt64(&alertCount) == 0 {
		t.Error("expected at least one alert to fire for expiring secret")
	}
}
