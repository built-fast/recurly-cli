#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# =============================================================================
# List Coupons
# =============================================================================

@test "coupons list default table output contains headers" {
  run "$RECURLY_BINARY" coupons list --limit 1
  assert_success
  assert_output_contains "Code"
  assert_output_contains "Name"
  assert_output_contains "Discount Type"
  assert_output_contains "State"
  assert_output_contains "Created At"
}

@test "coupons list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" coupons list --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "coupons list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" coupons list --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "coupons list --limit flag limits results" {
  run "$RECURLY_BINARY" coupons list --limit 1
  assert_success
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -ge 2 ]
}

@test "coupons list --order and --sort flags accepted" {
  run "$RECURLY_BINARY" coupons list --limit 1 --order desc --sort created_at
  assert_success
}

@test "coupons list --jq on list output extracts .object" {
  run "$RECURLY_BINARY" coupons list --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

# =============================================================================
# Get Coupon
# =============================================================================

@test "coupons get table detail output shows key-value pairs" {
  run "$RECURLY_BINARY" coupons get "code-coupon123"
  assert_success
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "Code"
  assert_output_contains "Name"
  assert_output_contains "State"
}

@test "coupons get --output json returns valid JSON object" {
  run "$RECURLY_BINARY" coupons get "code-coupon123" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "coupons get --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" coupons get "code-coupon123" --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "coupons get without coupon_id fails" {
  run "$RECURLY_BINARY" coupons get
  assert_failure
}

@test "coupons get --jq on single coupon extracts field" {
  run "$RECURLY_BINARY" coupons get "code-coupon123" --jq '.object'
  assert_success
  [ "$output" != "list" ]
  [ -n "$output" ]
}

# =============================================================================
# Create Percent Coupon
# =============================================================================

@test "coupons create-percent with required flags returns success" {
  run "$RECURLY_BINARY" coupons create-percent \
    --code "e2e_pct_coupon" \
    --name "E2E Percent Coupon" \
    --discount-percent 25
  assert_success
}

@test "coupons create-percent --output json returns valid JSON" {
  run "$RECURLY_BINARY" coupons create-percent \
    --code "e2e_pct_json" \
    --name "E2E Percent JSON" \
    --discount-percent 10 \
    --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "coupons create-percent missing --code fails" {
  run "$RECURLY_BINARY" coupons create-percent \
    --name "Missing Code" \
    --discount-percent 10
  assert_failure
}

@test "coupons create-percent missing --name fails" {
  run "$RECURLY_BINARY" coupons create-percent \
    --code "missing_name" \
    --discount-percent 10
  assert_failure
}

@test "coupons create-percent missing --discount-percent fails" {
  run "$RECURLY_BINARY" coupons create-percent \
    --code "missing_pct" \
    --name "Missing Percent"
  assert_failure
}

# =============================================================================
# Create Fixed Coupon
# =============================================================================

@test "coupons create-fixed with required flags returns success" {
  run "$RECURLY_BINARY" coupons create-fixed \
    --code "e2e_fixed_coupon" \
    --name "E2E Fixed Coupon" \
    --currency USD \
    --discount-amount 5.00
  assert_success
}

@test "coupons create-fixed --output json returns valid JSON" {
  run "$RECURLY_BINARY" coupons create-fixed \
    --code "e2e_fixed_json" \
    --name "E2E Fixed JSON" \
    --currency USD \
    --discount-amount 10.00 \
    --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "coupons create-fixed missing --currency fails" {
  run "$RECURLY_BINARY" coupons create-fixed \
    --code "missing_currency" \
    --name "Missing Currency" \
    --discount-amount 5.00
  assert_failure
}

@test "coupons create-fixed missing --discount-amount fails" {
  run "$RECURLY_BINARY" coupons create-fixed \
    --code "missing_amount" \
    --name "Missing Amount" \
    --currency USD
  assert_failure
}

# =============================================================================
# Create Free Trial Coupon
# =============================================================================

@test "coupons create-free-trial with required flags returns success" {
  run "$RECURLY_BINARY" coupons create-free-trial \
    --code "e2e_trial_coupon" \
    --name "E2E Trial Coupon" \
    --free-trial-amount 14 \
    --free-trial-unit "day"
  assert_success
}

@test "coupons create-free-trial --output json returns valid JSON" {
  run "$RECURLY_BINARY" coupons create-free-trial \
    --code "e2e_trial_json" \
    --name "E2E Trial JSON" \
    --free-trial-amount 30 \
    --free-trial-unit "day" \
    --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "coupons create-free-trial missing --free-trial-amount fails" {
  run "$RECURLY_BINARY" coupons create-free-trial \
    --code "missing_trial" \
    --name "Missing Trial Amount" \
    --free-trial-unit "day"
  assert_failure
}

@test "coupons create-free-trial missing --free-trial-unit fails" {
  run "$RECURLY_BINARY" coupons create-free-trial \
    --code "missing_unit" \
    --name "Missing Trial Unit" \
    --free-trial-amount 14
  assert_failure
}

# =============================================================================
# Update Coupon
# =============================================================================

@test "coupons update single field returns success" {
  run "$RECURLY_BINARY" coupons update "code-coupon123" --name "Updated Coupon Name"
  assert_success
}

@test "coupons update --output json returns valid JSON" {
  run "$RECURLY_BINARY" coupons update "code-coupon123" --name "Updated" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "coupons update without coupon_id fails" {
  run "$RECURLY_BINARY" coupons update
  assert_failure
}

# =============================================================================
# Deactivate Coupon
# =============================================================================

@test "coupons deactivate with --yes flag succeeds" {
  run "$RECURLY_BINARY" coupons deactivate "code-coupon123" --yes
  assert_success
}

@test "coupons deactivate without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" coupons deactivate \"code-coupon123\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "coupons deactivate without coupon_id fails" {
  run "$RECURLY_BINARY" coupons deactivate --yes
  assert_failure
}

@test "coupons deactivate --output json returns valid JSON" {
  run "$RECURLY_BINARY" coupons deactivate "code-coupon123" --yes --output json
  assert_success
  is_valid_json
}

# =============================================================================
# Restore Coupon
# =============================================================================

@test "coupons restore returns success" {
  run "$RECURLY_BINARY" coupons restore "code-coupon123"
  assert_success
}

@test "coupons restore with flags returns success" {
  run "$RECURLY_BINARY" coupons restore "code-coupon123" --name "Restored Coupon"
  assert_success
}

@test "coupons restore --output json returns valid JSON" {
  run "$RECURLY_BINARY" coupons restore "code-coupon123" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "coupons restore without coupon_id fails" {
  run "$RECURLY_BINARY" coupons restore
  assert_failure
}

# =============================================================================
# Generate Codes
# =============================================================================

@test "coupons generate-codes with required flags returns success" {
  run "$RECURLY_BINARY" coupons generate-codes "code-coupon123" --number-of-codes 5
  assert_success
}

@test "coupons generate-codes --output json returns valid JSON" {
  run "$RECURLY_BINARY" coupons generate-codes "code-coupon123" --number-of-codes 5 --output json
  assert_success
  is_valid_json
}

@test "coupons generate-codes missing --number-of-codes fails" {
  run "$RECURLY_BINARY" coupons generate-codes "code-coupon123"
  assert_failure
}

@test "coupons generate-codes without coupon_id fails" {
  run "$RECURLY_BINARY" coupons generate-codes --number-of-codes 5
  assert_failure
}

# =============================================================================
# List Codes
# =============================================================================

@test "coupons list-codes default table output contains headers" {
  run "$RECURLY_BINARY" coupons list-codes "code-coupon123" --limit 1
  assert_success
  assert_output_contains "Code"
  assert_output_contains "State"
  assert_output_contains "Created At"
}

@test "coupons list-codes --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" coupons list-codes "code-coupon123" --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "coupons list-codes --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" coupons list-codes "code-coupon123" --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "coupons list-codes --order and --sort flags accepted" {
  run "$RECURLY_BINARY" coupons list-codes "code-coupon123" --limit 1 --order desc --sort created_at
  assert_success
}

@test "coupons list-codes without coupon_id fails" {
  run "$RECURLY_BINARY" coupons list-codes
  assert_failure
}

# =============================================================================
# Help Tests
# =============================================================================

@test "coupons help text displays available commands" {
  run "$RECURLY_BINARY" coupons --help
  assert_success
  assert_output_contains "list"
  assert_output_contains "get"
  assert_output_contains "create-percent"
  assert_output_contains "create-fixed"
  assert_output_contains "create-free-trial"
  assert_output_contains "update"
  assert_output_contains "deactivate"
  assert_output_contains "restore"
  assert_output_contains "generate-codes"
  assert_output_contains "list-codes"
}

@test "coupons invalid subcommand shows help" {
  run "$RECURLY_BINARY" coupons notacommand
  assert_output_contains "Available Commands"
}
