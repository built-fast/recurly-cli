<p align="center">
    <img alt="BuiltFast Logo Light Mode" src="/assets/images/logo-light-mode.svg#gh-light-mode-only"/>
    <img alt="BuiltFast Logo Dark Mode" src="/assets/images/logo-dark-mode.svg#gh-dark-mode-only"/>
</p>

# recurly-cli

Command-line interface for the [Recurly](https://recurly.com) v3 API. Manage accounts, subscriptions, plans, invoices, items, coupons, and transactions from your terminal.

## Requirements

- Go 1.26+

## Install

```bash
go install github.com/built-fast/recurly-cli@latest
```

Or build from source:

```bash
make build    # produces ./bin/recurly
```

## Configuration

Run the interactive setup:

```bash
recurly configure
```

This creates `~/.config/recurly/config.toml` with your API key, region, and site.

You can also configure via environment variables:

```bash
export RECURLY_API_KEY=your-api-key
export RECURLY_REGION=us          # us (default) or eu
export RECURLY_SITE=mysite        # site subdomain
```

Or pass flags directly:

```bash
recurly accounts list --api-key <key> --region eu
```

Precedence: flags > environment variables > config file.

## Usage

```bash
# List resources
recurly accounts list --limit 50 --sort created_at --order desc

# Get a resource
recurly subscriptions get <subscription_id>

# Create
recurly accounts create --code acct-1 --email user@example.com

# Update
recurly subscriptions update <sub_id> --auto-renew true

# Create from file (JSON or YAML)
recurly plans create --from-file plan.yaml

# Destructive operations require confirmation (or --yes)
recurly subscriptions cancel <sub_id>
recurly subscriptions cancel <sub_id> --yes
```

### Resources

| Resource | Commands |
|---|---|
| `accounts` | list, get, create, update, deactivate, reactivate |
| `accounts billing-info` | get, update, remove |
| `accounts subscriptions` | list |
| `accounts invoices` | list |
| `accounts transactions` | list |
| `accounts redemptions` | list, list-active, create, remove |
| `subscriptions` | list, get, create, update, cancel, reactivate, pause, resume, terminate, convert-trial |
| `plans` | list, get, create, update, deactivate |
| `plans add-ons` | list, get, create, update, delete |
| `items` | list, get, create, update, deactivate, reactivate |
| `invoices` | list, get, line-items, collect, void, mark-failed |
| `coupons` | list, get, create-percent, create-fixed, create-free-trial, update, deactivate, restore, list-codes, generate-codes |
| `transactions` | list, get |

### Output

```bash
# Table (default), JSON, or pretty JSON
recurly accounts list --output table
recurly accounts list --output json
recurly accounts list --output json-pretty

# Select specific fields
recurly accounts list --field id,code,email

# Built-in jq filtering (no external jq needed)
recurly subscriptions list --jq '.data[] | select(.state == "active") | .id'
```

### Watch mode

Poll a resource on an interval:

```bash
recurly subscriptions get <sub_id> --watch 10s
```

### Open in browser

```bash
recurly open accounts <account_id>
```

### Shell completion

```bash
recurly completion bash
recurly completion zsh
```

## Development

Install dependencies:

```bash
brew bundle
```

Run the full check suite (formatting, linting, tests, e2e, vulnerability scan):

```bash
make check
```

Individual targets:

```bash
make build          # Build binary
make test           # Unit tests
make test-e2e       # E2E tests (BATS + Prism mock server)
make lint           # golangci-lint
make fmt            # Format code
make surface        # Regenerate CLI surface snapshot
make vuln           # Dependency vulnerability scan
```

## License

MIT
