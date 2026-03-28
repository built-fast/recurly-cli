#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# --- List Accounts ---

@test "accounts list returns success" {
  run "$RECURLY_BINARY" accounts list --limit 1
  assert_success
}

@test "accounts list --output json returns valid JSON" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json
  assert_success
  is_valid_json
}

@test "accounts list default output contains table headers" {
  run "$RECURLY_BINARY" accounts list --limit 1
  assert_success
  assert_output_contains "Code"
  assert_output_contains "Email"
  assert_output_contains "State"
}

# --- Get Account ---

@test "accounts get returns success for a valid account ID" {
  run "$RECURLY_BINARY" accounts get "code-test123"
  assert_success
}

@test "accounts get --output json returns valid JSON with expected fields" {
  run "$RECURLY_BINARY" accounts get "code-test123" --output json
  assert_success
  is_valid_json
}

@test "accounts get without account_id fails" {
  run "$RECURLY_BINARY" accounts get
  assert_failure
}

# --- Create Account ---

@test "accounts create returns success" {
  run "$RECURLY_BINARY" accounts create --code "e2e-test-acct"
  assert_success
}

@test "accounts create --output json returns valid JSON" {
  run "$RECURLY_BINARY" accounts create --code "e2e-json-acct" --output json
  assert_success
  is_valid_json
}

@test "accounts create with all flags returns success" {
  run "$RECURLY_BINARY" accounts create \
    --code "e2e-full-acct" \
    --email "test@example.com" \
    --first-name "Test" \
    --last-name "User" \
    --company "Test Corp"
  assert_success
}

# --- Update Account ---

@test "accounts update returns success" {
  run "$RECURLY_BINARY" accounts update "code-test123" --first-name "Updated"
  assert_success
}

@test "accounts update --output json returns valid JSON" {
  run "$RECURLY_BINARY" accounts update "code-test123" --email "updated@example.com" --output json
  assert_success
  is_valid_json
}

@test "accounts update without account_id fails" {
  run "$RECURLY_BINARY" accounts update
  assert_failure
}

# --- Deactivate Account ---

@test "accounts deactivate --yes returns success" {
  run "$RECURLY_BINARY" accounts deactivate "code-test123" --yes
  assert_success
}

@test "accounts deactivate --output json returns valid JSON" {
  run "$RECURLY_BINARY" accounts deactivate "code-test123" --yes --output json
  assert_success
  is_valid_json
}

@test "accounts deactivate without account_id fails" {
  run "$RECURLY_BINARY" accounts deactivate --yes
  assert_failure
}

# --- Reactivate Account ---

@test "accounts reactivate --yes returns success" {
  run "$RECURLY_BINARY" accounts reactivate "code-test123" --yes
  assert_success
}

@test "accounts reactivate --output json returns valid JSON" {
  run "$RECURLY_BINARY" accounts reactivate "code-test123" --yes --output json
  assert_success
  is_valid_json
}

@test "accounts reactivate without account_id fails" {
  run "$RECURLY_BINARY" accounts reactivate --yes
  assert_failure
}
