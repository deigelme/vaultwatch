package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full VaultWatch configuration.
type Config struct {
	Vault         VaultConfig    `yaml:"vault"`
	Secrets       []SecretConfig `yaml:"secrets"`
	CheckInterval time.Duration  `yaml:"check_interval"`
	WarnBefore    time.Duration  `yaml:"warn_before"`
	CriticalBefore time.Duration `yaml:"critical_before"`
	Alerts        []AlertConfig  `yaml:"alerts"`
}

// VaultConfig holds Vault connection settings.
type VaultConfig struct {
	Address     string `yaml:"address"`
	Token       string `yaml:"token"`
	TLSSkipVerify bool `yaml:"tls_skip_verify"`
	CACert      string `yaml:"ca_cert"`
}

// SecretConfig describes a single secret path to monitor.
type SecretConfig struct {
	Path           string        `yaml:"path"`
	WarnBefore     time.Duration `yaml:"warn_before"`
	CriticalBefore time.Duration `yaml:"critical_before"`
}

// AlertConfig holds the type and arbitrary key/value settings for a notifier.
type AlertConfig struct {
	Type string `yaml:"type"`
	// All remaining fields are decoded into Extra for notifier-specific use.
	Extra map[string]string `yaml:",inline"`
}

const (
	defaultCheckInterval  = 5 * time.Minute
	defaultWarnBefore     = 72 * time.Hour
	defaultCriticalBefore = 24 * time.Hour
)

// Load reads and validates a YAML configuration file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse %q: %w", path, err)
	}

	if cfg.Vault.Address == "" {
		return nil, fmt.Errorf("config: vault.address is required")
	}
	if len(cfg.Secrets) == 0 {
		return nil, fmt.Errorf("config: at least one secret path is required")
	}
	if cfg.CheckInterval <= 0 {
		cfg.CheckInterval = defaultCheckInterval
	}
	if cfg.WarnBefore <= 0 {
		cfg.WarnBefore = defaultWarnBefore
	}
	if cfg.CriticalBefore <= 0 {
		cfg.CriticalBefore = defaultCriticalBefore
	}
	return &cfg, nil
}
