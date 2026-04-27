# vaultwatch

A CLI tool that monitors HashiCorp Vault secret expiration and sends configurable alerts before rotation deadlines.

---

## Installation

```bash
go install github.com/yourusername/vaultwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/vaultwatch.git
cd vaultwatch && go build -o vaultwatch .
```

---

## Usage

Configure your Vault address and alert thresholds, then run:

```bash
export VAULT_ADDR="https://vault.example.com"
export VAULT_TOKEN="s.your-token-here"

vaultwatch watch --threshold 7d --alert slack
```

Check a specific secret path:

```bash
vaultwatch check secret/data/myapp/db-credentials
```

List all secrets approaching expiration:

```bash
vaultwatch list --threshold 30d --output table
```

### Common Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--threshold` | Alert window before expiration | `14d` |
| `--alert` | Alert backend (`slack`, `email`, `webhook`) | `stdout` |
| `--interval` | Polling interval | `1h` |
| `--output` | Output format (`table`, `json`) | `table` |

---

## Configuration

VaultWatch can be configured via a `vaultwatch.yaml` file:

```yaml
vault_addr: https://vault.example.com
threshold: 7d
alert:
  type: slack
  webhook_url: https://hooks.slack.com/services/...
```

---

## License

MIT © 2024 yourusername