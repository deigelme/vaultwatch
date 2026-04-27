package vault_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yourusername/vaultwatch/internal/vault"
)

func newMockVaultServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	return httptest.NewServer(handler)
}

func TestNewClient_ValidConfig(t *testing.T) {
	client, err := vault.NewClient("http://127.0.0.1:8200", "test-token")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if client == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestNewClient_InvalidAddress(t *testing.T) {
	// Vault API client creation itself won't fail on bad address;
	// errors surface on actual requests. Just ensure no panic.
	client, err := vault.NewClient("://bad-url", "token")
	if err == nil && client == nil {
		t.Fatal("expected either an error or a client")
	}
}

func TestGetSecretMeta_NotFound(t *testing.T) {
	server := newMockVaultServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{}`))
	})
	defer server.Close()

	client, err := vault.NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = client.GetSecretMeta("secret/data/missing")
	if err == nil {
		t.Fatal("expected error for missing secret, got nil")
	}
}

func TestGetSecretMeta_Success(t *testing.T) {
	server := newMockVaultServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"lease_id": "secret/data/myapp/token/abc123",
			"renewable": true,
			"lease_duration": 3600,
			"data": {"value": "s3cr3t"}
		}`))
	})
	defer server.Close()

	client, err := vault.NewClient(server.URL, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	meta, err := client.GetSecretMeta("secret/data/myapp/token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if meta.TTL.Seconds() != 3600 {
		t.Errorf("expected TTL 3600s, got %v", meta.TTL)
	}
	if !meta.Renewable {
		t.Error("expected secret to be renewable")
	}
}
