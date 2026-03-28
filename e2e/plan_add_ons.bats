#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# =============================================================================
# List Plan Add-Ons
# =============================================================================

@test "plans add-ons list default table output contains headers" {
  run "$RECURLY_BINARY" plans add-ons list "code-plan123" --limit 1
  assert_success
  assert_output_contains "ID"
  assert_output_contains "Code"
  assert_output_contains "Name"
  assert_output_contains "State"
  assert_output_contains "Add-On Type"
  assert_output_contains "Created At"
}

@test "plans add-ons list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" plans add-ons list "code-plan123" --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "plans add-ons list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" plans add-ons list "code-plan123" --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "plans add-ons list --state flag accepted" {
  run "$RECURLY_BINARY" plans add-ons list "code-plan123" --limit 1 --state active
  assert_success
}

@test "plans add-ons list --order and --sort flags accepted" {
  run "$RECURLY_BINARY" plans add-ons list "code-plan123" --limit 1 --order desc --sort created_at
  assert_success
}

@test "plans add-ons list --jq extracts .object" {
  run "$RECURLY_BINARY" plans add-ons list "code-plan123" --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

@test "plans add-ons list without plan_id fails" {
  run "$RECURLY_BINARY" plans add-ons list
  assert_failure
}

# =============================================================================
# Get Plan Add-On
# =============================================================================

@test "plans add-ons get table detail output shows key-value pairs" {
  run "$RECURLY_BINARY" plans add-ons get "code-plan123" "test-addon"
  assert_success
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "Code"
  assert_output_contains "Name"
  assert_output_contains "State"
}

@test "plans add-ons get --output json returns valid JSON object" {
  run "$RECURLY_BINARY" plans add-ons get "code-plan123" "test-addon" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "plans add-ons get --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" plans add-ons get "code-plan123" "test-addon" --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "plans add-ons get without args fails" {
  run "$RECURLY_BINARY" plans add-ons get
  assert_failure
}

@test "plans add-ons get with only plan_id fails" {
  run "$RECURLY_BINARY" plans add-ons get "code-plan123"
  assert_failure
}

# =============================================================================
# Create Plan Add-On
# =============================================================================

@test "plans add-ons create with minimal flags returns success" {
  run "$RECURLY_BINARY" plans add-ons create "code-plan123" \
    --code "e2e-addon" \
    --name "E2E Add-On" \
    --currency USD \
    --unit-amount 5.00
  assert_success
}

@test "plans add-ons create --output json returns valid JSON" {
  run "$RECURLY_BINARY" plans add-ons create "code-plan123" \
    --code "e2e-addon-json" \
    --name "E2E JSON Add-On" \
    --currency USD \
    --unit-amount 5.00 \
    --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "plans add-ons create without plan_id fails" {
  run "$RECURLY_BINARY" plans add-ons create
  assert_failure
}

# =============================================================================
# Update Plan Add-On
# =============================================================================

@test "plans add-ons update single field returns success" {
  run "$RECURLY_BINARY" plans add-ons update "code-plan123" "test-addon" --name "Updated Add-On"
  assert_success
}

@test "plans add-ons update --output json returns valid JSON" {
  run "$RECURLY_BINARY" plans add-ons update "code-plan123" "test-addon" --name "Updated" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "plans add-ons update without args fails" {
  run "$RECURLY_BINARY" plans add-ons update
  assert_failure
}

# =============================================================================
# Delete Plan Add-On
# =============================================================================

@test "plans add-ons delete with --yes flag succeeds" {
  run "$RECURLY_BINARY" plans add-ons delete "code-plan123" "test-addon" --yes
  assert_success
}

@test "plans add-ons delete without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" plans add-ons delete \"code-plan123\" \"test-addon\" 2>&1"
  assert_success
  assert_output_contains "cancelled"
}

@test "plans add-ons delete without args fails" {
  run "$RECURLY_BINARY" plans add-ons delete --yes
  assert_failure
}

# =============================================================================
# Help Tests
# =============================================================================

@test "plans add-ons help text displays available commands" {
  run "$RECURLY_BINARY" plans add-ons --help
  assert_success
  assert_output_contains "list"
  assert_output_contains "get"
  assert_output_contains "create"
  assert_output_contains "update"
  assert_output_contains "delete"
}
