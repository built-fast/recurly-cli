#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# =============================================================================
# List Plans
# =============================================================================

@test "plans list default table output contains headers" {
  run "$RECURLY_BINARY" plans list --limit 1
  assert_success
  assert_output_contains "Code"
  assert_output_contains "Name"
  assert_output_contains "State"
  assert_output_contains "Interval"
  assert_output_contains "Price"
  assert_output_contains "Created At"
}

@test "plans list --output json returns valid JSON with envelope" {
  run "$RECURLY_BINARY" plans list --limit 1 --output json
  assert_success
  is_valid_json
  assert_json_value ".object" "list"
  # Verify .data is an array
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  [ "$data_type" = "array" ]
}

@test "plans list --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" plans list --limit 1 --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "plans list --limit flag limits results" {
  run "$RECURLY_BINARY" plans list --limit 1
  assert_success
  # Should get output (at least header + 1 row)
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -ge 2 ]
}

@test "plans list --state flag accepted" {
  run "$RECURLY_BINARY" plans list --limit 1 --state active
  assert_success
}

@test "plans list --order and --sort flags accepted" {
  run "$RECURLY_BINARY" plans list --limit 1 --order desc --sort created_at
  assert_success
}

@test "plans list --jq on list output extracts .object" {
  run "$RECURLY_BINARY" plans list --limit 1 --jq '.object'
  assert_success
  [ "$output" = "list" ]
}

# =============================================================================
# Get Plan
# =============================================================================

@test "plans get table detail output shows key-value pairs" {
  run "$RECURLY_BINARY" plans get "code-plan123"
  assert_success
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "Code"
  assert_output_contains "Name"
  assert_output_contains "State"
}

@test "plans get --output json returns valid JSON object" {
  run "$RECURLY_BINARY" plans get "code-plan123" --output json
  assert_success
  is_valid_json
  # Should be an object, not a list envelope
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "plans get --output json-pretty returns indented JSON" {
  run "$RECURLY_BINARY" plans get "code-plan123" --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 1 ]
}

@test "plans get without plan_id fails" {
  run "$RECURLY_BINARY" plans get
  assert_failure
}

@test "plans get --jq on single plan extracts field" {
  run "$RECURLY_BINARY" plans get "code-plan123" --jq '.object'
  assert_success
  # Should be a resource object type (e.g. "plan"), not "list"
  [ "$output" != "list" ]
  [ -n "$output" ]
}

# =============================================================================
# Create Plan
# =============================================================================

@test "plans create with minimal flags returns success" {
  run "$RECURLY_BINARY" plans create \
    --code "e2e_plan_min" \
    --name "E2E Minimal Plan" \
    --interval-unit "months" \
    --interval-length 1 \
    --currency USD \
    --unit-amount 9.99
  assert_success
}

@test "plans create with description returns success" {
  run "$RECURLY_BINARY" plans create \
    --code "e2e_plan_desc" \
    --name "E2E Desc Plan" \
    --interval-unit "months" \
    --interval-length 1 \
    --currency USD \
    --unit-amount 19.99 \
    --description "A test plan with description"
  assert_success
}

@test "plans create with trial settings returns success" {
  run "$RECURLY_BINARY" plans create \
    --code "e2e_plan_trial" \
    --name "E2E Trial Plan" \
    --interval-unit "months" \
    --interval-length 1 \
    --currency USD \
    --unit-amount 29.99 \
    --trial-unit "days" \
    --trial-length 14
  assert_success
}

@test "plans create --output json returns valid JSON" {
  run "$RECURLY_BINARY" plans create \
    --code "e2e_plan_json" \
    --name "E2E JSON Plan" \
    --interval-unit "months" \
    --interval-length 1 \
    --currency USD \
    --unit-amount 9.99 \
    --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

# =============================================================================
# Update Plan
# =============================================================================

@test "plans update single field returns success" {
  run "$RECURLY_BINARY" plans update "code-plan123" --name "Updated Plan Name"
  assert_success
}

@test "plans update --output json returns valid JSON" {
  run "$RECURLY_BINARY" plans update "code-plan123" --name "Updated" --output json
  assert_success
  is_valid_json
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  [ "$obj_type" = "object" ]
}

@test "plans update without plan_id fails" {
  run "$RECURLY_BINARY" plans update
  assert_failure
}

# =============================================================================
# Deactivate Plan
# =============================================================================

@test "plans deactivate with --yes flag succeeds" {
  run "$RECURLY_BINARY" plans deactivate "code-plan123" --yes
  assert_success
}

@test "plans deactivate without --yes piping 'n' cancels" {
  run bash -c "echo 'n' | \"$RECURLY_BINARY\" plans deactivate \"code-plan123\" 2>&1"
  # The command exits 0 on cancellation (prints message to stderr)
  assert_success
  assert_output_contains "cancelled"
}

@test "plans deactivate without plan_id fails" {
  run "$RECURLY_BINARY" plans deactivate --yes
  assert_failure
}

# =============================================================================
# Error / Help Tests
# =============================================================================

@test "plans invalid subcommand shows help" {
  run "$RECURLY_BINARY" plans notacommand
  # Cobra shows help text for unknown subcommands
  assert_output_contains "Available Commands"
}

@test "plans help text displays available commands" {
  run "$RECURLY_BINARY" plans --help
  assert_success
  assert_output_contains "list"
  assert_output_contains "get"
  assert_output_contains "create"
  assert_output_contains "update"
  assert_output_contains "deactivate"
}
