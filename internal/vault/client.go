package vault

import (
	"fmt"
	"time"

	vaultapi "github.com/hashicorp/vault/api"
)

// Client wraps the Vault API client with helper methods.
type Client struct {
	api *vaultapi.Client
}

// SecretMeta holds metadata about a Vault secret relevant to expiration.
type SecretMeta struct {
	Path       string
	ExpiresAt  time.Time
	TTL        time.Duration
	Renewable  bool
	LeaseID    string
}

// NewClient creates a new Vault client using the provided address and token.
func NewClient(address, token string) (*Client, error) {
	cfg := vaultapi.DefaultConfig()
	cfg.Address = address

	api, err := vaultapi.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("creating vault api client: %w", err)
	}

	api.SetToken(token)

	return &Client{api: api}, nil
}

// GetSecretMeta reads a KV secret path and returns its metadata.
func (c *Client) GetSecretMeta(path string) (*SecretMeta, error) {
	secret, err := c.api.Logical().Read(path)
	if err != nil {
		return nil, fmt.Errorf("reading secret at %q: %w", path, err)
	}
	if secret == nil {
		return nil, fmt.Errorf("secret not found at path %q", path)
	}

	ttl := time.Duration(secret.LeaseDuration) * time.Second
	expiresAt := time.Now().Add(ttl)

	return &SecretMeta{
		Path:      path,
		ExpiresAt: expiresAt,
		TTL:       ttl,
		Renewable: secret.Renewable,
		LeaseID:   secret.LeaseID,
	}, nil
}

// HealthCheck verifies connectivity to the Vault server.
func (c *Client) HealthCheck() error {
	_, err := c.api.Sys().Health()
	if err != nil {
		return fmt.Errorf("vault health check failed: %w", err)
	}
	return nil
}
