#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# =============================================================================
# List Redemptions
# =============================================================================

@test "accounts redemptions list returns success" {
  run "$RECURLY_BINARY" accounts redemptions list "code-test123" --limit 1
  assert_success
}

@test "accounts redemptions list default table output contains headers" {
  run "$RECURLY_BINARY" accounts redemptions list "code-test123" --limit 1
  assert_success
  assert_output_contains "ID"
  assert_output_contains "Coupon Code"
  assert_output_contains "State"
  assert_output_contains "Currency"
  assert_output_contains "Created At"
}

@test "accounts redemptions list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" accounts redemptions list "code-test123" --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "accounts redemptions list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" accounts redemptions list "code-test123" --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "accounts redemptions list --jq extracts .object" {
  run "$RECURLY_BINARY" accounts redemptions list "code-test123" --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

@test "accounts redemptions list without account_id fails" {
  run "$RECURLY_BINARY" accounts redemptions list
  assert_failure
}

# =============================================================================
# List Active Redemptions
# Prism confuses /coupon_redemptions/active with /coupon_redemptions/{id},
# so API-dependent tests are skipped. Argument validation still tested.
# =============================================================================

@test "accounts redemptions list-active returns success" {
  skip "Prism routing conflict: /active matched as /{coupon_redemption_id}"
  run "$RECURLY_BINARY" accounts redemptions list-active "code-test123"
  assert_success
}

@test "accounts redemptions list-active --output json returns valid JSON with envelope" {
  skip "Prism routing conflict: /active matched as /{coupon_redemption_id}"
  run "$RECURLY_BINARY" accounts redemptions list-active "code-test123" --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
}

@test "accounts redemptions list-active without account_id fails" {
  run "$RECURLY_BINARY" accounts redemptions list-active
  assert_failure
}

# =============================================================================
# Create Redemption
# =============================================================================

@test "accounts redemptions create with required flags returns success" {
  run "$RECURLY_BINARY" accounts redemptions create "code-test123" \
    --coupon-id "coupon123"
  assert_success
}

@test "accounts redemptions create table output shows key-value pairs" {
  run "$RECURLY_BINARY" accounts redemptions create "code-test123" \
    --coupon-id "coupon123"
  assert_success
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "ID"
  assert_output_contains "State"
}

@test "accounts redemptions create --output json returns valid JSON" {
  run "$RECURLY_BINARY" accounts redemptions create "code-test123" \
    --coupon-id "coupon123" \
    --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "accounts redemptions create with optional flags returns success" {
  run "$RECURLY_BINARY" accounts redemptions create "code-test123" \
    --coupon-id "coupon123" \
    --currency "USD"
  assert_success
}

@test "accounts redemptions create missing --coupon-id fails" {
  run "$RECURLY_BINARY" accounts redemptions create "code-test123"
  assert_failure
}

@test "accounts redemptions create without account_id fails" {
  run "$RECURLY_BINARY" accounts redemptions create
  assert_failure
}

# =============================================================================
# Remove Redemption
# =============================================================================

@test "accounts redemptions remove with --yes returns success" {
  run "$RECURLY_BINARY" accounts redemptions remove "code-test123" --yes
  assert_success
}

@test "accounts redemptions remove --yes --output json returns valid JSON" {
  run "$RECURLY_BINARY" accounts redemptions remove "code-test123" --yes --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "accounts redemptions remove --yes specific redemption returns success" {
  run "$RECURLY_BINARY" accounts redemptions remove "code-test123" "redemption123" --yes
  assert_success
}

@test "accounts redemptions remove without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" accounts redemptions remove \"code-test123\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "accounts redemptions remove without --yes piping 'y' confirms" {
  run bash -c "echo 'y' | \"$RECURLY_BINARY\" accounts redemptions remove \"code-test123\" 2>&1"
  assert_success
  assert_output_not_contains "cancelled"
}

@test "accounts redemptions remove without account_id fails" {
  run "$RECURLY_BINARY" accounts redemptions remove --yes
  assert_failure
}

# =============================================================================
# Help Tests
# =============================================================================

@test "accounts redemptions help text displays available commands" {
  run "$RECURLY_BINARY" accounts redemptions --help
  assert_success
  assert_output_contains "list"
  assert_output_contains "list-active"
  assert_output_contains "create"
  assert_output_contains "remove"
}
