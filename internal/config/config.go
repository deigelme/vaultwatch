package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds the full vaultwatch configuration.
type Config struct {
	Vault   VaultConfig   `yaml:"vault"`
	Alerts  AlertsConfig  `yaml:"alerts"`
	Monitor MonitorConfig `yaml:"monitor"`
}

// VaultConfig contains Vault connection settings.
type VaultConfig struct {
	Address   string `yaml:"address"`
	Token     string `yaml:"token"`
	Namespace string `yaml:"namespace"`
}

// AlertsConfig defines how and when alerts are sent.
type AlertsConfig struct {
	SlackWebhook string          `yaml:"slack_webhook"`
	Email        EmailConfig     `yaml:"email"`
	Thresholds   []time.Duration `yaml:"thresholds"`
}

// EmailConfig holds SMTP settings for email alerts.
type EmailConfig struct {
	SMTPHost   string   `yaml:"smtp_host"`
	SMTPPort   int      `yaml:"smtp_port"`
	From       string   `yaml:"from"`
	Recipients []string `yaml:"recipients"`
}

// MonitorConfig controls the monitoring behaviour.
type MonitorConfig struct {
	Interval   time.Duration `yaml:"interval"`
	SecretPaths []string     `yaml:"secret_paths"`
}

// Load reads and parses a YAML config file from the given path.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parsing config file %q: %w", path, err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &cfg, nil
}

// validate performs basic sanity checks on the loaded configuration.
func (c *Config) validate() error {
	if c.Vault.Address == "" {
		return fmt.Errorf("vault.address is required")
	}
	if c.Vault.Token == "" {
		return fmt.Errorf("vault.token is required")
	}
	if len(c.Monitor.SecretPaths) == 0 {
		return fmt.Errorf("monitor.secret_paths must contain at least one path")
	}
	if c.Monitor.Interval <= 0 {
		c.Monitor.Interval = 5 * time.Minute
	}
	if len(c.Alerts.Thresholds) == 0 {
		c.Alerts.Thresholds = []time.Duration{72 * time.Hour, 24 * time.Hour}
	}
	return nil
}
