#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# =============================================================================
# Account Subscriptions List
# =============================================================================

@test "accounts subscriptions list returns success" {
  run "$RECURLY_BINARY" accounts subscriptions list "code-test123" --limit 1
  assert_success
}

@test "accounts subscriptions list default table output contains headers" {
  run "$RECURLY_BINARY" accounts subscriptions list "code-test123" --limit 1
  assert_success
  assert_output_contains "ID"
  assert_output_contains "Plan"
  assert_output_contains "State"
  assert_output_contains "Currency"
}

@test "accounts subscriptions list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" accounts subscriptions list "code-test123" --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "accounts subscriptions list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" accounts subscriptions list "code-test123" --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "accounts subscriptions list --state flag accepted" {
  run "$RECURLY_BINARY" accounts subscriptions list "code-test123" --limit 1 --state active
  assert_success
}

@test "accounts subscriptions list --jq extracts .object" {
  run "$RECURLY_BINARY" accounts subscriptions list "code-test123" --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

@test "accounts subscriptions list without account_id fails" {
  run "$RECURLY_BINARY" accounts subscriptions list
  assert_failure
}

# =============================================================================
# Account Invoices List
# =============================================================================

@test "accounts invoices list returns success" {
  run "$RECURLY_BINARY" accounts invoices list "code-test123" --limit 1
  assert_success
}

@test "accounts invoices list default table output contains headers" {
  run "$RECURLY_BINARY" accounts invoices list "code-test123" --limit 1
  assert_success
  assert_output_contains "State"
  assert_output_contains "Type"
  assert_output_contains "Currency"
  assert_output_contains "Created At"
}

@test "accounts invoices list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" accounts invoices list "code-test123" --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "accounts invoices list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" accounts invoices list "code-test123" --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "accounts invoices list --type flag accepted" {
  run "$RECURLY_BINARY" accounts invoices list "code-test123" --limit 1 --type charge
  assert_success
}

@test "accounts invoices list --jq extracts .object" {
  run "$RECURLY_BINARY" accounts invoices list "code-test123" --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

@test "accounts invoices list without account_id fails" {
  run "$RECURLY_BINARY" accounts invoices list
  assert_failure
}

# =============================================================================
# Account Transactions List
# =============================================================================

@test "accounts transactions list returns success" {
  run "$RECURLY_BINARY" accounts transactions list "code-test123" --limit 1
  assert_success
}

@test "accounts transactions list default table output contains headers" {
  run "$RECURLY_BINARY" accounts transactions list "code-test123" --limit 1
  assert_success
  assert_output_contains "ID"
  assert_output_contains "Type"
  assert_output_contains "Amount"
  assert_output_contains "Currency"
  assert_output_contains "Status"
}

@test "accounts transactions list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" accounts transactions list "code-test123" --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "accounts transactions list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" accounts transactions list "code-test123" --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "accounts transactions list --type flag accepted" {
  run "$RECURLY_BINARY" accounts transactions list "code-test123" --limit 1 --type payment
  assert_success
}

@test "accounts transactions list --jq extracts .object" {
  run "$RECURLY_BINARY" accounts transactions list "code-test123" --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

@test "accounts transactions list without account_id fails" {
  run "$RECURLY_BINARY" accounts transactions list
  assert_failure
}

# =============================================================================
# Help Tests
# =============================================================================

@test "accounts subscriptions help displays list command" {
  run "$RECURLY_BINARY" accounts subscriptions --help
  assert_success
  assert_output_contains "list"
}

@test "accounts invoices help displays list command" {
  run "$RECURLY_BINARY" accounts invoices --help
  assert_success
  assert_output_contains "list"
}

@test "accounts transactions help displays list command" {
  run "$RECURLY_BINARY" accounts transactions --help
  assert_success
  assert_output_contains "list"
}
