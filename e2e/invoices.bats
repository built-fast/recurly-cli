#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# =============================================================================
# List Invoices
# =============================================================================

@test "invoices list returns success" {
  run "$RECURLY_BINARY" invoices list --limit 1
  assert_success
}

@test "invoices list default table output contains headers" {
  run "$RECURLY_BINARY" invoices list --limit 1
  assert_success
  assert_output_contains "ID"
  assert_output_contains "Number"
  assert_output_contains "Type"
  assert_output_contains "State"
  assert_output_contains "Currency"
  assert_output_contains "Total"
  assert_output_contains "Created At"
}

@test "invoices list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" invoices list --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "invoices list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" invoices list --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "invoices list --state flag accepted" {
  run "$RECURLY_BINARY" invoices list --limit 1 --state pending
  assert_success
}

@test "invoices list --type flag accepted" {
  run "$RECURLY_BINARY" invoices list --limit 1 --type charge
  assert_success
}

@test "invoices list --order and --sort flags accepted" {
  run "$RECURLY_BINARY" invoices list --limit 1 --order desc --sort created_at
  assert_success
}

@test "invoices list --jq extracts .object" {
  run "$RECURLY_BINARY" invoices list --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

# =============================================================================
# Get Invoice
# =============================================================================

@test "invoices get returns success" {
  run "$RECURLY_BINARY" invoices get "inv-test123"
  assert_success
}

@test "invoices get table output shows key-value pairs" {
  run "$RECURLY_BINARY" invoices get "inv-test123"
  assert_success
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "State"
  assert_output_contains "Type"
}

@test "invoices get --output json returns valid JSON object" {
  run "$RECURLY_BINARY" invoices get "inv-test123" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "invoices get --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" invoices get "inv-test123" --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "invoices get without invoice_id fails" {
  run "$RECURLY_BINARY" invoices get
  assert_failure
}

@test "invoices get --jq extracts field" {
  run "$RECURLY_BINARY" invoices get "inv-test123" --jq '.object'
  assert_success
  [ "$output" != "list" ]
  [ -n "$output" ]
}

# =============================================================================
# Help / Error Tests
# =============================================================================

@test "invoices help text displays available commands" {
  run "$RECURLY_BINARY" invoices --help
  assert_success
  assert_output_contains "list"
  assert_output_contains "get"
  assert_output_contains "void"
  assert_output_contains "collect"
  assert_output_contains "mark-failed"
  assert_output_contains "line-items"
}

@test "invoices invalid subcommand shows help" {
  run "$RECURLY_BINARY" invoices notacommand
  assert_output_contains "Available Commands"
}
