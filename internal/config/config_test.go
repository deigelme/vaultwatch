package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func writeTempConfig(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	p := filepath.Join(dir, "vaultwatch.yaml")
	if err := os.WriteFile(p, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return p
}

func TestLoad_ValidConfig(t *testing.T) {
	p := writeTempConfig(t, `
vault:
  address: http://127.0.0.1:8200
  token: root
secrets:
  - path: secret/myapp/db
    warn_before: 168h
    crit_before: 24h
alerts:
  - type: stdout
interval: 1m
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Vault.Address != "http://127.0.0.1:8200" {
		t.Errorf("unexpected vault address: %s", cfg.Vault.Address)
	}
	if len(cfg.Secrets) != 1 {
		t.Fatalf("expected 1 secret, got %d", len(cfg.Secrets))
	}
	if cfg.Secrets[0].Path != "secret/myapp/db" {
		t.Errorf("unexpected secret path: %s", cfg.Secrets[0].Path)
	}
	if cfg.Interval != time.Minute {
		t.Errorf("unexpected interval: %v", cfg.Interval)
	}
}

func TestLoad_MissingVaultAddress(t *testing.T) {
	p := writeTempConfig(t, `
vault:
  token: root
secrets:
  - path: secret/myapp/db
`)
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for missing vault address")
	}
}

func TestLoad_NoSecretPaths(t *testing.T) {
	p := writeTempConfig(t, `
vault:
  address: http://127.0.0.1:8200
  token: root
`)
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error when no secrets defined")
	}
}

func TestLoad_DefaultInterval(t *testing.T) {
	p := writeTempConfig(t, `
vault:
  address: http://127.0.0.1:8200
  token: root
secrets:
  - path: secret/myapp/db
`)
	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Interval != 5*time.Minute {
		t.Errorf("expected default interval 5m, got %v", cfg.Interval)
	}
	if cfg.Secrets[0].WarnBefore != 7*24*time.Hour {
		t.Errorf("expected default warn_before 168h, got %v", cfg.Secrets[0].WarnBefore)
	}
	if cfg.Secrets[0].CritBefore != 24*time.Hour {
		t.Errorf("expected default crit_before 24h, got %v", cfg.Secrets[0].CritBefore)
	}
}
