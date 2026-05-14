package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the top-level vaultwatch configuration.
type Config struct {
	Vault   VaultConfig    `yaml:"vault"`
	Secrets []SecretConfig `yaml:"secrets"`
	Alerts  []AlertConfig  `yaml:"alerts"`
	Interval time.Duration `yaml:"interval"`
}

// VaultConfig holds Vault connection settings.
type VaultConfig struct {
	Address string `yaml:"address"`
	Token   string `yaml:"token"`
	RoleID  string `yaml:"role_id"`
	SecretID string `yaml:"secret_id"`
}

// SecretConfig describes a single secret to monitor.
type SecretConfig struct {
	Path        string        `yaml:"path"`
	WarnBefore  time.Duration `yaml:"warn_before"`
	CritBefore  time.Duration `yaml:"crit_before"`
}

// AlertConfig holds a notifier type and its arbitrary options.
type AlertConfig struct {
	Type    string                 `yaml:"type"`
	Options map[string]interface{} `yaml:",inline"`
}

// Load reads and validates a vaultwatch config file at path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("config: read file: %w", err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("config: parse yaml: %w", err)
	}

	if cfg.Vault.Address == "" {
		return nil, fmt.Errorf("config: vault.address is required")
	}
	if len(cfg.Secrets) == 0 {
		return nil, fmt.Errorf("config: at least one secret path is required")
	}
	if cfg.Interval == 0 {
		cfg.Interval = 5 * time.Minute
	}
	for i, s := range cfg.Secrets {
		if s.WarnBefore == 0 {
			cfg.Secrets[i].WarnBefore = 7 * 24 * time.Hour
		}
		if s.CritBefore == 0 {
			cfg.Secrets[i].CritBefore = 24 * time.Hour
		}
	}
	return &cfg, nil
}
