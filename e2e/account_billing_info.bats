#!/usr/bin/env bats

load "test_helper"

# =============================================================================
# Get Billing Info
# =============================================================================

@test "accounts billing-info get returns success" {
  run "$RECURLY_BINARY" accounts billing-info get "code-test123"
  assert_success
}

@test "accounts billing-info get table output shows key-value pairs" {
  run "$RECURLY_BINARY" accounts billing-info get "code-test123"
  assert_success
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "Valid"
}

@test "accounts billing-info get --output json returns valid JSON object" {
  run "$RECURLY_BINARY" accounts billing-info get "code-test123" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "accounts billing-info get --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" accounts billing-info get "code-test123" --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "accounts billing-info get without account_id fails" {
  run "$RECURLY_BINARY" accounts billing-info get
  assert_failure
}

# =============================================================================
# Update Billing Info
# =============================================================================

@test "accounts billing-info update single field returns success" {
  run "$RECURLY_BINARY" accounts billing-info update "code-test123" --first-name "Updated"
  assert_success
}

@test "accounts billing-info update --output json returns valid JSON" {
  run "$RECURLY_BINARY" accounts billing-info update "code-test123" --first-name "Updated" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "accounts billing-info update without account_id fails" {
  run "$RECURLY_BINARY" accounts billing-info update
  assert_failure
}

# =============================================================================
# Remove Billing Info
# =============================================================================

@test "accounts billing-info remove with --yes flag succeeds" {
  run "$RECURLY_BINARY" accounts billing-info remove "code-test123" --yes
  assert_success
  assert_output_contains "Billing info removed"
}

@test "accounts billing-info remove without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" accounts billing-info remove \"code-test123\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "accounts billing-info remove without account_id fails" {
  run "$RECURLY_BINARY" accounts billing-info remove --yes
  assert_failure
}

# =============================================================================
# Help Tests
# =============================================================================

@test "accounts billing-info help text displays available commands" {
  run "$RECURLY_BINARY" accounts billing-info --help
  assert_success
  assert_output_contains "get"
  assert_output_contains "update"
  assert_output_contains "remove"
}
