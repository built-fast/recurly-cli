#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues.

# =============================================================================
# --field / -f Flag Tests
# Tests verify field selection on list and get commands.
# =============================================================================

# --- Basic Field Selection (List) ---

@test "accounts list --field code shows only Code column" {
  run "$RECURLY_BINARY" accounts list --limit 1 --field code
  assert_success
  assert_output_contains "Code"
  assert_output_not_contains "Email"
  assert_output_not_contains "First Name"
}

@test "accounts list --field code,email shows only selected columns" {
  run "$RECURLY_BINARY" accounts list --limit 1 --field code,email
  assert_success
  assert_output_contains "Code"
  assert_output_contains "Email"
  assert_output_not_contains "First Name"
  assert_output_not_contains "Company"
}

@test "accounts list -f (short flag) works" {
  run "$RECURLY_BINARY" accounts list --limit 1 -f code
  assert_success
  assert_output_contains "Code"
  assert_output_not_contains "Email"
}

# --- Field Selection with JSON Output ---

@test "accounts list --field code --output json filters JSON keys" {
  run "$RECURLY_BINARY" accounts list --limit 1 --field code --output json
  assert_success
  is_valid_json
  # The data items should have the code field
  local has_code
  has_code="$(echo "$output" | jq '.data[0] | has("code")')"
  [ "$has_code" = "true" ]
  # Should not have email field
  local has_email
  has_email="$(echo "$output" | jq '.data[0] | has("email")')"
  [ "$has_email" = "false" ]
}

# --- Field Selection (Get) ---

@test "accounts get --field code shows only Code field" {
  run "$RECURLY_BINARY" accounts get "code-test123" --field code
  assert_success
  assert_output_contains "Code"
  assert_output_not_contains "Email"
}

@test "accounts get --field code --output json filters JSON keys" {
  run "$RECURLY_BINARY" accounts get "code-test123" --field code --output json
  assert_success
  is_valid_json
  local has_code
  has_code="$(echo "$output" | jq 'has("code")')"
  [ "$has_code" = "true" ]
  local has_email
  has_email="$(echo "$output" | jq 'has("email")')"
  [ "$has_email" = "false" ]
}

# --- Case Insensitive ---

@test "accounts list --field CODE is case insensitive" {
  run "$RECURLY_BINARY" accounts list --limit 1 --field CODE
  assert_success
  assert_output_contains "Code"
}

# --- Invalid Fields ---

@test "accounts list --field nonexistent_field fails with available fields" {
  run "$RECURLY_BINARY" accounts list --limit 1 --field nonexistent_field
  assert_failure
  assert_output_contains "unknown field"
}

# --- Multiple Fields Preserve Order ---

@test "accounts list --field email,code shows fields in specified order" {
  run "$RECURLY_BINARY" accounts list --limit 1 --field email,code
  assert_success
  assert_output_contains "Email"
  assert_output_contains "Code"
}
