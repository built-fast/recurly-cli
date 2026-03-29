#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# =============================================================================
# List Transactions
# =============================================================================

@test "transactions list returns success" {
  run "$RECURLY_BINARY" transactions list --limit 1
  assert_success
}

@test "transactions list default table output contains headers" {
  run "$RECURLY_BINARY" transactions list --limit 1
  assert_success
  assert_output_contains "ID"
  assert_output_contains "Type"
  assert_output_contains "Status"
  assert_output_contains "Currency"
  assert_output_contains "Amount"
  assert_output_contains "Created At"
}

@test "transactions list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" transactions list --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "transactions list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" transactions list --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "transactions list --type flag accepted" {
  run "$RECURLY_BINARY" transactions list --limit 1 --type payment
  assert_success
}

@test "transactions list --order and --sort flags accepted" {
  run "$RECURLY_BINARY" transactions list --limit 1 --order desc --sort created_at
  assert_success
}

@test "transactions list --jq extracts .object" {
  run "$RECURLY_BINARY" transactions list --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

# =============================================================================
# Get Transaction
# =============================================================================

@test "transactions get returns success" {
  run "$RECURLY_BINARY" transactions get "txn-test123"
  assert_success
}

@test "transactions get table output shows key-value pairs" {
  run "$RECURLY_BINARY" transactions get "txn-test123"
  assert_success
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "Type"
  assert_output_contains "Status"
}

@test "transactions get --output json returns valid JSON object" {
  run "$RECURLY_BINARY" transactions get "txn-test123" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "transactions get --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" transactions get "txn-test123" --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "transactions get without transaction_id fails" {
  run "$RECURLY_BINARY" transactions get
  assert_failure
}

@test "transactions get --jq extracts field" {
  run "$RECURLY_BINARY" transactions get "txn-test123" --jq '.object'
  assert_success
  [ "$output" != "list" ]
  [ -n "$output" ]
}

# =============================================================================
# Help / Error Tests
# =============================================================================

@test "transactions help text displays available commands" {
  run "$RECURLY_BINARY" transactions --help
  assert_success
  assert_output_contains "list"
  assert_output_contains "get"
}

@test "transactions invalid subcommand shows help" {
  run "$RECURLY_BINARY" transactions notacommand
  assert_output_contains "Available Commands"
}
