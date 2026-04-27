package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/yourusername/vaultwatch/internal/config"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "vaultwatch-*.yaml")
	if err != nil {
		t.Fatalf("creating temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("writing temp file: %v", err)
	}
	f.Close()
	return f.Name()
}

func TestLoad_ValidConfig(t *testing.T) {
	content := `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
monitor:
  interval: 10m
  secret_paths:
    - secret/myapp/db
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("expected vault address, got %q", cfg.Vault.Address)
	}
	if cfg.Monitor.Interval != 10*time.Minute {
		t.Errorf("expected 10m interval, got %v", cfg.Monitor.Interval)
	}
	// Default thresholds should be applied
	if len(cfg.Alerts.Thresholds) != 2 {
		t.Errorf("expected 2 default thresholds, got %d", len(cfg.Alerts.Thresholds))
	}
}

func TestLoad_MissingVaultAddress(t *testing.T) {
	content := `
vault:
  token: "root"
monitor:
  secret_paths:
    - secret/myapp/db
`
	path := writeTempConfig(t, content)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for missing vault address, got nil")
	}
}

func TestLoad_NoSecretPaths(t *testing.T) {
	content := `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
monitor:
  secret_paths: []
`
	path := writeTempConfig(t, content)
	_, err := config.Load(path)
	if err == nil {
		t.Fatal("expected error for empty secret_paths, got nil")
	}
}

func TestLoad_DefaultInterval(t *testing.T) {
	content := `
vault:
  address: "http://127.0.0.1:8200"
  token: "root"
monitor:
  secret_paths:
    - secret/myapp/db
`
	path := writeTempConfig(t, content)
	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Monitor.Interval != 5*time.Minute {
		t.Errorf("expected default 5m interval, got %v", cfg.Monitor.Interval)
	}
}
