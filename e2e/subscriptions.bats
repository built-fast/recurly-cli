#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# =============================================================================
# List Subscriptions
# =============================================================================

@test "subscriptions list returns success" {
  run "$RECURLY_BINARY" subscriptions list --limit 1
  assert_success
}

@test "subscriptions list default table output contains headers" {
  run "$RECURLY_BINARY" subscriptions list --limit 1
  assert_success
  assert_output_contains "ID"
  assert_output_contains "Account"
  assert_output_contains "Plan"
  assert_output_contains "State"
  assert_output_contains "Currency"
}

@test "subscriptions list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" subscriptions list --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "subscriptions list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" subscriptions list --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "subscriptions list --state flag accepted" {
  run "$RECURLY_BINARY" subscriptions list --limit 1 --state active
  assert_success
}

@test "subscriptions list --order and --sort flags accepted" {
  run "$RECURLY_BINARY" subscriptions list --limit 1 --order desc --sort created_at
  assert_success
}

@test "subscriptions list --jq extracts .object" {
  run "$RECURLY_BINARY" subscriptions list --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

# =============================================================================
# Get Subscription
# =============================================================================

@test "subscriptions get returns success" {
  run "$RECURLY_BINARY" subscriptions get "sub-test123"
  assert_success
}

@test "subscriptions get table output shows key-value pairs" {
  run "$RECURLY_BINARY" subscriptions get "sub-test123"
  assert_success
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "State"
}

@test "subscriptions get --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions get "sub-test123" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "subscriptions get --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" subscriptions get "sub-test123" --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "subscriptions get without subscription_id fails" {
  run "$RECURLY_BINARY" subscriptions get
  assert_failure
}

# =============================================================================
# Create Subscription
# =============================================================================

@test "subscriptions create with minimal flags returns success" {
  run "$RECURLY_BINARY" subscriptions create \
    --plan-code "test-plan" \
    --account-code "test-account" \
    --currency "USD"
  assert_success
}

@test "subscriptions create --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions create \
    --plan-code "test-plan" \
    --account-code "test-account" \
    --currency "USD" \
    --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "subscriptions create with additional flags returns success" {
  run "$RECURLY_BINARY" subscriptions create \
    --plan-code "test-plan" \
    --account-code "test-account" \
    --currency "USD" \
    --quantity 2 \
    --unit-amount 9.99 \
    --collection-method "automatic"
  assert_success
}

# =============================================================================
# Update Subscription
# =============================================================================

@test "subscriptions update returns success" {
  run "$RECURLY_BINARY" subscriptions update "sub-test123" --collection-method "manual"
  assert_success
}

@test "subscriptions update --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions update "sub-test123" --collection-method "manual" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "subscriptions update without subscription_id fails" {
  run "$RECURLY_BINARY" subscriptions update
  assert_failure
}

# =============================================================================
# Cancel Subscription
# =============================================================================

@test "subscriptions cancel --yes returns success" {
  run "$RECURLY_BINARY" subscriptions cancel "sub-test123" --yes
  assert_success
}

@test "subscriptions cancel --yes --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions cancel "sub-test123" --yes --output json
  assert_success
  is_valid_json
}

@test "subscriptions cancel --yes --timeframe bill_date returns success" {
  run "$RECURLY_BINARY" subscriptions cancel "sub-test123" --yes --timeframe bill_date
  assert_success
}

@test "subscriptions cancel --yes --timeframe term_end returns success" {
  run "$RECURLY_BINARY" subscriptions cancel "sub-test123" --yes --timeframe term_end
  assert_success
}

@test "subscriptions cancel without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" subscriptions cancel \"sub-test123\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "subscriptions cancel without subscription_id fails" {
  run "$RECURLY_BINARY" subscriptions cancel --yes
  assert_failure
}

# =============================================================================
# Reactivate Subscription
# =============================================================================

@test "subscriptions reactivate --yes returns success" {
  run "$RECURLY_BINARY" subscriptions reactivate "sub-test123" --yes
  assert_success
}

@test "subscriptions reactivate --yes --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions reactivate "sub-test123" --yes --output json
  assert_success
  is_valid_json
}

@test "subscriptions reactivate without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" subscriptions reactivate \"sub-test123\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "subscriptions reactivate without subscription_id fails" {
  run "$RECURLY_BINARY" subscriptions reactivate --yes
  assert_failure
}

# =============================================================================
# Pause Subscription
# =============================================================================

@test "subscriptions pause --yes with --remaining-pause-cycles returns success" {
  run "$RECURLY_BINARY" subscriptions pause "sub-test123" --yes --remaining-pause-cycles 3
  assert_success
}

@test "subscriptions pause --yes --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions pause "sub-test123" --yes --remaining-pause-cycles 2 --output json
  assert_success
  is_valid_json
}

@test "subscriptions pause --yes --remaining-pause-cycles flag variations" {
  run "$RECURLY_BINARY" subscriptions pause "sub-test123" --yes --remaining-pause-cycles 1
  assert_success
  run "$RECURLY_BINARY" subscriptions pause "sub-test123" --yes --remaining-pause-cycles 5
  assert_success
}

@test "subscriptions pause without --remaining-pause-cycles fails" {
  run "$RECURLY_BINARY" subscriptions pause "sub-test123" --yes --no-input
  assert_failure
}

@test "subscriptions pause without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" subscriptions pause \"sub-test123\" --remaining-pause-cycles 3 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "subscriptions pause without subscription_id fails" {
  run "$RECURLY_BINARY" subscriptions pause --yes --remaining-pause-cycles 3
  assert_failure
}

# =============================================================================
# Resume Subscription
# =============================================================================

@test "subscriptions resume --yes returns success" {
  run "$RECURLY_BINARY" subscriptions resume "sub-test123" --yes
  assert_success
}

@test "subscriptions resume --yes --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions resume "sub-test123" --yes --output json
  assert_success
  is_valid_json
}

@test "subscriptions resume without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" subscriptions resume \"sub-test123\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "subscriptions resume without subscription_id fails" {
  run "$RECURLY_BINARY" subscriptions resume --yes
  assert_failure
}

# =============================================================================
# Terminate Subscription
# =============================================================================

@test "subscriptions terminate --yes returns success" {
  run "$RECURLY_BINARY" subscriptions terminate "sub-test123" --yes
  assert_success
}

@test "subscriptions terminate --yes --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions terminate "sub-test123" --yes --output json
  assert_success
  is_valid_json
}

@test "subscriptions terminate --yes --refund full returns success" {
  run "$RECURLY_BINARY" subscriptions terminate "sub-test123" --yes --refund full
  assert_success
}

@test "subscriptions terminate --yes --refund partial returns success" {
  run "$RECURLY_BINARY" subscriptions terminate "sub-test123" --yes --refund partial
  assert_success
}

@test "subscriptions terminate --yes --refund none returns success" {
  run "$RECURLY_BINARY" subscriptions terminate "sub-test123" --yes --refund none
  assert_success
}

@test "subscriptions terminate --yes --charge flag returns success" {
  run "$RECURLY_BINARY" subscriptions terminate "sub-test123" --yes --charge
  assert_success
}

@test "subscriptions terminate without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" subscriptions terminate \"sub-test123\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "subscriptions terminate without subscription_id fails" {
  run "$RECURLY_BINARY" subscriptions terminate --yes
  assert_failure
}

# =============================================================================
# Convert Trial Subscription
# =============================================================================

@test "subscriptions convert-trial --yes returns success" {
  run "$RECURLY_BINARY" subscriptions convert-trial "sub-test123" --yes
  assert_success
}

@test "subscriptions convert-trial --yes --output json returns valid JSON" {
  run "$RECURLY_BINARY" subscriptions convert-trial "sub-test123" --yes --output json
  assert_success
  is_valid_json
}

@test "subscriptions convert-trial without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" subscriptions convert-trial \"sub-test123\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "subscriptions convert-trial without subscription_id fails" {
  run "$RECURLY_BINARY" subscriptions convert-trial --yes
  assert_failure
}

# =============================================================================
# Help / Error Tests
# =============================================================================

@test "subscriptions help text displays all available commands" {
  run "$RECURLY_BINARY" subscriptions --help
  assert_success
  assert_output_contains "list"
  assert_output_contains "get"
  assert_output_contains "create"
  assert_output_contains "update"
  assert_output_contains "cancel"
  assert_output_contains "reactivate"
  assert_output_contains "pause"
  assert_output_contains "resume"
  assert_output_contains "terminate"
  assert_output_contains "convert-trial"
}

@test "subscriptions invalid subcommand shows help" {
  run "$RECURLY_BINARY" subscriptions notacommand
  assert_output_contains "Available Commands"
}
