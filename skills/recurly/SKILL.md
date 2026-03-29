---
name: recurly
description: Recurly CLI — manage Recurly billing resources (accounts, subscriptions, plans, invoices, items, coupons, transactions) from the command line. Use this skill when the user needs to interact with the Recurly API.
triggers:
  - recurly
  - billing
  - subscription management
  - invoice
  - recurring billing
---

# Recurly CLI Skill

The `recurly` CLI manages Recurly billing resources from the terminal. This
document teaches agents how to invoke any command correctly.

## Authentication

Precedence (highest to lowest):

1. `--api-key` flag on any command
2. `RECURLY_API_KEY` environment variable
3. Config file (created via `recurly configure`)

```bash
# Interactive setup (writes ~/.config/recurly/config.toml)
recurly configure

# One-off override
recurly accounts list --api-key <key>

# Environment variable
export RECURLY_API_KEY=<key>
recurly accounts list
```

## Region

Default: `us`. Set via `--region`, `RECURLY_REGION` env, or `recurly configure`.

```bash
recurly accounts list --region eu
```

## Output Modes

| Flag | Effect |
|---|---|
| `--output table` | Human-readable table (default for TTY) |
| `--output json` | Compact JSON |
| `--output json-pretty` | Indented JSON |
| `--jq <expr>` | Apply jq expression to JSON output (built-in, no external jq) |
| `--field <fields>` | Comma-separated field names to display |

Agent invariant: always use `--output json` when parsing output programmatically.
The `--jq` flag implies JSON output. `--jq` and `--output table` are mutually
exclusive.

### Field Selection

```bash
recurly accounts list --field id,code,email
recurly accounts get acct123 --field id,email --output json
```

## Exit Codes

| Code | Meaning |
|---|---|
| 0 | Success |
| 1 | General error (invalid flags, validation failure, network error, server error, auth error, not found) |

Error messages are written to stderr. Agents should check exit code and parse
stderr for diagnostics.

## Pagination

List commands support these flags:

| Flag | Type | Description |
|---|---|---|
| `--limit` | int | Max results to return (default 20) |
| `--all` | bool | Fetch every page (overrides --limit) |
| `--sort` | string | Sort field (e.g., `created_at`, `updated_at`) |
| `--order` | string | Sort order (`asc` or `desc`) |

JSON list output envelope:

```json
{
  "object": "list",
  "has_more": true,
  "data": [...]
}
```

Use `--all` to iterate all pages. Use `--limit` to cap results.

## Destructive Operations

Commands that delete, deactivate, void, cancel, or terminate require
confirmation. Use `--yes` to skip the interactive prompt in scripts:

```bash
recurly accounts deactivate acct123 --yes
recurly subscriptions cancel sub456 --yes
```

## File-Based Input

Create and update commands accept `--from-file` (`-F`) for complex payloads.
Supports JSON and YAML files, or `-` for stdin. CLI flags override file values.

```bash
recurly accounts create --from-file account.json
recurly plans create -F plan.yaml
echo '{"code":"new-plan","name":"New"}' | recurly plans create --from-file -
```

## Watch Mode

Get commands support `--watch` to poll for changes:

```bash
recurly subscriptions get sub123 --watch        # default 5s interval
recurly subscriptions get sub123 --watch 30s    # custom interval
```

Terminal mode: clears and refreshes in place. Piped mode: emits
newline-delimited output. Ctrl+C to stop.

## Browser Navigation

```bash
recurly open                              # Open dashboard home
recurly open accounts                     # Open accounts list
recurly open subscriptions sub123         # Open specific subscription
recurly open --url accounts acct456       # Print URL without opening
recurly open --site mysite accounts       # Override site subdomain
```

Supported resources: accounts, plans, subscriptions, invoices, transactions,
items, coupons.

## Command Reference

### recurly configure

Interactive setup for API key, region, and site subdomain.

```
recurly configure
```

### Accounts

#### recurly accounts list

List accounts with optional filters.

```
recurly accounts list [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--email`, `--subscriber`,
`--past-due`, `--begin-time`, `--end-time`

#### recurly accounts get

Get a single account by ID.

```
recurly accounts get <account_id> [flags]
```

Flags: `--watch`

#### recurly accounts create

Create an account.

```
recurly accounts create [flags]
```

Flags: `--code`, `--email`, `--first-name`, `--last-name`, `--company`,
`--bill-to`, `--preferred-locale`, `--tax-exempt`, `--vat-number`, `--from-file`

#### recurly accounts update

Update an account.

```
recurly accounts update <account_id> [flags]
```

Flags: `--email`, `--first-name`, `--last-name`, `--company`, `--bill-to`,
`--preferred-locale`, `--tax-exempt`, `--vat-number`, `--from-file`

#### recurly accounts deactivate

Deactivate an account.

```
recurly accounts deactivate <account_id> [flags]
```

Flags: `--yes`

#### recurly accounts reactivate

Reactivate a deactivated account.

```
recurly accounts reactivate <account_id> [flags]
```

Flags: `--yes`

### Account Billing Info

#### recurly accounts billing-info get

Get billing info for an account.

```
recurly accounts billing-info get <account_id> [flags]
```

Flags: `--watch`

#### recurly accounts billing-info update

Update billing info for an account.

```
recurly accounts billing-info update <account_id> [flags]
```

Flags: `--first-name`, `--last-name`, `--company`, `--vat-number`, `--currency`,
`--token-id`, `--address-street1`, `--address-street2`, `--address-city`,
`--address-region`, `--address-country`, `--address-postal-code`,
`--primary-payment-method`, `--backup-payment-method`, `--from-file`

#### recurly accounts billing-info remove

Remove billing info from an account.

```
recurly accounts billing-info remove <account_id> [flags]
```

Flags: `--yes`

### Account Subscriptions

#### recurly accounts subscriptions list

List subscriptions for an account.

```
recurly accounts subscriptions list <account_id> [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--state`

### Account Invoices

#### recurly accounts invoices list

List invoices for an account.

```
recurly accounts invoices list <account_id> [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--state`, `--type`

### Account Transactions

#### recurly accounts transactions list

List transactions for an account.

```
recurly accounts transactions list <account_id> [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--type`, `--success`

### Account Coupon Redemptions

#### recurly accounts redemptions list

List coupon redemptions for an account.

```
recurly accounts redemptions list <account_id> [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`

#### recurly accounts redemptions list-active

List active coupon redemptions for an account.

```
recurly accounts redemptions list-active <account_id> [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`

#### recurly accounts redemptions create

Create a coupon redemption on an account.

```
recurly accounts redemptions create <account_id> [flags]
```

Flags: `--coupon-id`, `--currency`, `--subscription-id`, `--no-input`,
`--from-file`

#### recurly accounts redemptions remove

Remove a coupon redemption from an account.

```
recurly accounts redemptions remove <account_id> [redemption_id] [flags]
```

Flags: `--yes`

### Subscriptions

#### recurly subscriptions list

List subscriptions with optional filters.

```
recurly subscriptions list [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--state`, `--plan-id`,
`--begin-time`, `--end-time`

#### recurly subscriptions get

Get a single subscription by ID.

```
recurly subscriptions get <subscription_id> [flags]
```

Flags: `--watch`

#### recurly subscriptions create

Create a subscription.

```
recurly subscriptions create [flags]
```

Flags: `--plan-code`, `--account-code`, `--currency`, `--quantity`,
`--unit-amount`, `--collection-method`, `--auto-renew`, `--net-terms`,
`--net-terms-type`, `--po-number`, `--coupon-code`, `--billing-info-id`,
`--gateway-code`, `--starts-at`, `--next-bill-date`, `--trial-ends-at`,
`--total-billing-cycles`, `--renewal-billing-cycles`, `--from-file`

#### recurly subscriptions update

Update a subscription.

```
recurly subscriptions update <subscription_id> [flags]
```

Flags: `--collection-method`, `--auto-renew`, `--net-terms`, `--net-terms-type`,
`--po-number`, `--billing-info-id`, `--gateway-code`, `--next-bill-date`,
`--remaining-billing-cycles`, `--renewal-billing-cycles`,
`--revenue-schedule-type`, `--customer-notes`, `--terms-and-conditions`,
`--from-file`

#### recurly subscriptions cancel

Cancel a subscription.

```
recurly subscriptions cancel <subscription_id> [flags]
```

Flags: `--timeframe`, `--yes`

#### recurly subscriptions reactivate

Reactivate a canceled subscription.

```
recurly subscriptions reactivate <subscription_id> [flags]
```

Flags: `--yes`

#### recurly subscriptions pause

Pause a subscription.

```
recurly subscriptions pause <subscription_id> [flags]
```

Flags: `--remaining-pause-cycles`, `--no-input`, `--yes`

#### recurly subscriptions resume

Resume a paused subscription.

```
recurly subscriptions resume <subscription_id> [flags]
```

Flags: `--yes`

#### recurly subscriptions terminate

Terminate a subscription immediately.

```
recurly subscriptions terminate <subscription_id> [flags]
```

Flags: `--refund`, `--charge`, `--yes`

#### recurly subscriptions convert-trial

Convert a trial subscription.

```
recurly subscriptions convert-trial <subscription_id> [flags]
```

Flags: `--yes`

### Plans

#### recurly plans list

List plans.

```
recurly plans list [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--state`

#### recurly plans get

Get a single plan by ID.

```
recurly plans get <plan_id> [flags]
```

Flags: `--watch`

#### recurly plans create

Create a plan.

```
recurly plans create [flags]
```

Flags: `--code`, `--name`, `--description`, `--interval-length`,
`--interval-unit`, `--currency`, `--unit-amount`, `--setup-fee`,
`--accounting-code`, `--tax-code`, `--tax-exempt`, `--auto-renew`,
`--total-billing-cycles`, `--trial-length`, `--trial-unit`,
`--trial-requires-billing-info`, `--pricing-model`, `--display-quantity`,
`--allow-any-item-on-subscriptions`, `--bypass-confirmation`,
`--revenue-schedule-type`, `--setup-fee-revenue-schedule-type`,
`--setup-fee-accounting-code`, `--avalara-transaction-type`,
`--avalara-service-type`, `--vertex-transaction-type`,
`--harmonized-system-code`, `--dunning-campaign-id`, `--cancel-url`,
`--success-url`, `--liability-gl-account-id`, `--revenue-gl-account-id`,
`--performance-obligation-id`, `--setup-fee-liability-gl-account-id`,
`--setup-fee-revenue-gl-account-id`, `--setup-fee-performance-obligation-id`,
`--from-file`

#### recurly plans update

Update a plan.

```
recurly plans update <plan_id> [flags]
```

Flags: `--code`, `--name`, `--description`, `--currency`, `--unit-amount`,
`--setup-fee`, `--accounting-code`, `--tax-code`, `--tax-exempt`, `--auto-renew`,
`--total-billing-cycles`, `--trial-length`, `--trial-unit`,
`--trial-requires-billing-info`, `--display-quantity`,
`--allow-any-item-on-subscriptions`, `--bypass-confirmation`,
`--revenue-schedule-type`, `--setup-fee-revenue-schedule-type`,
`--setup-fee-accounting-code`, `--avalara-transaction-type`,
`--avalara-service-type`, `--vertex-transaction-type`,
`--harmonized-system-code`, `--dunning-campaign-id`, `--cancel-url`,
`--success-url`, `--liability-gl-account-id`, `--revenue-gl-account-id`,
`--performance-obligation-id`, `--setup-fee-liability-gl-account-id`,
`--setup-fee-revenue-gl-account-id`, `--setup-fee-performance-obligation-id`,
`--from-file`

#### recurly plans deactivate

Deactivate a plan.

```
recurly plans deactivate <plan_id> [flags]
```

Flags: `--yes`

### Plan Add-Ons

#### recurly plans add-ons list

List add-ons for a plan.

```
recurly plans add-ons list <plan_id> [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--state`

#### recurly plans add-ons get

Get a single add-on.

```
recurly plans add-ons get <plan_id> <add_on_id> [flags]
```

Flags: `--watch`

#### recurly plans add-ons create

Create an add-on for a plan.

```
recurly plans add-ons create <plan_id> [flags]
```

Flags: `--code`, `--name`, `--add-on-type`, `--unit-amount`, `--currency`,
`--default-quantity`, `--display-quantity`, `--optional`, `--accounting-code`,
`--tax-code`, `--revenue-schedule-type`, `--usage-type`,
`--usage-calculation-type`, `--measured-unit-id`, `--from-file`

#### recurly plans add-ons update

Update an add-on.

```
recurly plans add-ons update <plan_id> <add_on_id> [flags]
```

Flags: `--code`, `--name`, `--unit-amount`, `--currency`, `--default-quantity`,
`--display-quantity`, `--optional`, `--accounting-code`, `--tax-code`,
`--revenue-schedule-type`, `--usage-calculation-type`, `--measured-unit-id`,
`--measured-unit-name`, `--from-file`

#### recurly plans add-ons delete

Delete an add-on from a plan.

```
recurly plans add-ons delete <plan_id> <add_on_id> [flags]
```

Flags: `--yes`

### Items

#### recurly items list

List items.

```
recurly items list [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--state`, `--begin-time`,
`--end-time`

#### recurly items get

Get a single item by ID.

```
recurly items get <item_id> [flags]
```

Flags: `--watch`

#### recurly items create

Create an item.

```
recurly items create [flags]
```

Flags: `--code`, `--name`, `--description`, `--external-sku`, `--currency`,
`--unit-amount`, `--accounting-code`, `--tax-code`, `--tax-exempt`,
`--revenue-schedule-type`, `--avalara-transaction-type`,
`--avalara-service-type`, `--harmonized-system-code`, `--from-file`

#### recurly items update

Update an item.

```
recurly items update <item_id> [flags]
```

Flags: `--code`, `--name`, `--description`, `--external-sku`, `--currency`,
`--unit-amount`, `--accounting-code`, `--tax-code`, `--tax-exempt`,
`--revenue-schedule-type`, `--avalara-transaction-type`,
`--avalara-service-type`, `--harmonized-system-code`, `--from-file`

#### recurly items deactivate

Deactivate an item.

```
recurly items deactivate <item_id> [flags]
```

Flags: `--yes`

#### recurly items reactivate

Reactivate an item.

```
recurly items reactivate <item_id> [flags]
```

Flags: `--yes`

### Invoices

#### recurly invoices list

List invoices.

```
recurly invoices list [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--state`, `--type`,
`--begin-time`, `--end-time`

#### recurly invoices get

Get a single invoice by ID.

```
recurly invoices get <invoice_id> [flags]
```

Flags: `--line-items`, `--watch`

#### recurly invoices line-items

List line items for an invoice.

```
recurly invoices line-items <invoice_id> [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--begin-time`, `--end-time`

#### recurly invoices collect

Collect payment on a past-due invoice.

```
recurly invoices collect <invoice_id> [flags]
```

Flags: `--yes`

#### recurly invoices void

Void an invoice.

```
recurly invoices void <invoice_id> [flags]
```

Flags: `--yes`

#### recurly invoices mark-failed

Mark an invoice as failed.

```
recurly invoices mark-failed <invoice_id> [flags]
```

Flags: `--yes`

### Transactions

#### recurly transactions list

List transactions.

```
recurly transactions list [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--type`, `--success`,
`--begin-time`, `--end-time`

#### recurly transactions get

Get a single transaction by ID.

```
recurly transactions get <transaction_id> [flags]
```

Flags: `--watch`

### Coupons

#### recurly coupons list

List coupons.

```
recurly coupons list [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--begin-time`, `--end-time`

#### recurly coupons get

Get a single coupon by ID.

```
recurly coupons get <coupon_id> [flags]
```

Flags: `--watch`

#### recurly coupons create-percent

Create a percent-based coupon.

```
recurly coupons create-percent [flags]
```

Flags: `--code`, `--name`, `--discount-percent`, `--coupon-type`, `--duration`,
`--temporal-amount`, `--temporal-unit`, `--applies-to-all-plans`,
`--applies-to-all-items`, `--applies-to-non-plan-charges`, `--plan-codes`,
`--item-codes`, `--max-redemptions`, `--max-redemptions-per-account`,
`--redeem-by`, `--redemption-resource`, `--invoice-description`,
`--hosted-page-description`, `--unique-code-template`, `--no-input`,
`--from-file`

#### recurly coupons create-fixed

Create a fixed-amount coupon.

```
recurly coupons create-fixed [flags]
```

Flags: `--code`, `--name`, `--currency`, `--discount-amount`, `--coupon-type`,
`--duration`, `--temporal-amount`, `--temporal-unit`, `--applies-to-all-plans`,
`--applies-to-all-items`, `--applies-to-non-plan-charges`, `--plan-codes`,
`--item-codes`, `--max-redemptions`, `--max-redemptions-per-account`,
`--redeem-by`, `--redemption-resource`, `--invoice-description`,
`--hosted-page-description`, `--unique-code-template`, `--no-input`,
`--from-file`

#### recurly coupons create-free-trial

Create a free-trial coupon.

```
recurly coupons create-free-trial [flags]
```

Flags: `--code`, `--name`, `--free-trial-amount`, `--free-trial-unit`,
`--coupon-type`, `--applies-to-all-plans`, `--plan-codes`, `--max-redemptions`,
`--max-redemptions-per-account`, `--redeem-by`, `--redemption-resource`,
`--invoice-description`, `--hosted-page-description`, `--unique-code-template`,
`--no-input`, `--from-file`

#### recurly coupons update

Update a coupon.

```
recurly coupons update <coupon_id> [flags]
```

Flags: `--name`, `--max-redemptions`, `--max-redemptions-per-account`,
`--redeem-by-date`, `--invoice-description`, `--hosted-description`,
`--from-file`

#### recurly coupons deactivate

Deactivate a coupon.

```
recurly coupons deactivate <coupon_id> [flags]
```

Flags: `--yes`

#### recurly coupons restore

Restore a deactivated coupon.

```
recurly coupons restore <coupon_id> [flags]
```

Flags: `--name`, `--max-redemptions`, `--max-redemptions-per-account`,
`--redeem-by-date`, `--invoice-description`, `--hosted-description`

#### recurly coupons list-codes

List unique coupon codes for a coupon.

```
recurly coupons list-codes <coupon_id> [flags]
```

Flags: `--limit`, `--all`, `--sort`, `--order`, `--redeemed`

#### recurly coupons generate-codes

Generate unique coupon codes.

```
recurly coupons generate-codes <coupon_id> [flags]
```

Flags: `--number-of-codes`, `--no-input`

### Utility Commands

#### recurly open

Open Recurly dashboard in browser.

```
recurly open [resource] [identifier] [flags]
```

Flags: `--site`, `--url`

#### recurly skill

Print this skill reference document to stdout.

```
recurly skill
```

#### recurly skill install

Install SKILL.md to the Claude Code skills directory.

```
recurly skill install
```

#### recurly skill uninstall

Remove installed SKILL.md from the skills directories.

```
recurly skill uninstall
```

## Common Agent Workflows

### Create a Resource

```bash
# Simple: use flags
recurly accounts create --code acct-1 --email user@example.com --first-name Jane --last-name Doe

# Complex: use --from-file
recurly plans create --from-file plan.json --output json
```

### List and Filter Resources

```bash
# List with pagination
recurly accounts list --limit 50 --sort created_at --order desc --output json

# Filter with jq
recurly subscriptions list --all --output json --jq '.data[] | select(.state == "active") | .id'

# Select specific fields
recurly accounts list --field id,code,email --limit 10
```

### Update a Resource

```bash
recurly accounts update acct-1 --email new@example.com --output json
recurly subscriptions update sub123 --from-file changes.yaml --output json
```

### Deactivate / Cancel / Terminate

```bash
# Always pass --yes for non-interactive use
recurly accounts deactivate acct-1 --yes
recurly subscriptions cancel sub123 --timeframe end_of_term --yes
recurly subscriptions terminate sub456 --refund full --yes
```

### Monitor a Resource

```bash
recurly subscriptions get sub123 --watch 10s --output json
```

### Open Dashboard

```bash
recurly open accounts acct-1
```
