#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# --- Default Table Output (List) ---

@test "default output for list command contains table headers" {
  run "$RECURLY_BINARY" accounts list --limit 1
  assert_success
  assert_output_contains "Code"
  assert_output_contains "Email"
  assert_output_contains "First Name"
  assert_output_contains "Last Name"
  assert_output_contains "Company"
  assert_output_contains "State"
  assert_output_contains "Created At"
}

@test "default table list output contains data rows" {
  run "$RECURLY_BINARY" accounts list --limit 1
  assert_success
  # Table output should have at least 2 lines: header + data row(s)
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  if [ "$line_count" -lt 2 ]; then
    echo "expected at least 2 lines (header + data), got $line_count" >&2
    echo "output: $output" >&2
    return 1
  fi
}

# --- JSON Output (List) ---

@test "--output json for list produces compact valid JSON" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json
  assert_success
  is_valid_json
  # Compact JSON should be a single line (no indentation)
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  if [ "$line_count" -ne 1 ]; then
    echo "expected compact JSON on 1 line, got $line_count lines" >&2
    echo "output: $output" >&2
    return 1
  fi
}

@test "--output json for list produces a JSON array" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json
  assert_success
  is_valid_json
  # Verify it's an array
  local arr_type
  arr_type="$(echo "$output" | jq -r 'type')"
  if [ "$arr_type" != "array" ]; then
    echo "expected JSON array, got $arr_type" >&2
    return 1
  fi
}

@test "--output json for list contains expected top-level fields" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json
  assert_success
  is_valid_json
  # Verify expected account fields exist in the first element
  local has_code
  has_code="$(echo "$output" | jq '.[0] | has("code")')"
  if [ "$has_code" != "true" ]; then
    echo "expected JSON objects to have 'code' field" >&2
    echo "output: $output" >&2
    return 1
  fi
}

# --- JSON Pretty Output (List) ---

@test "--output json-pretty for list produces indented valid JSON" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json-pretty
  assert_success
  is_valid_json
  # Pretty JSON should have multiple lines when results exist.
  # An empty array "[]" is valid but single-line — only assert multi-line
  # when Prism returns data.
  if [ "$output" != "[]" ]; then
    local line_count
    line_count="$(echo "$output" | wc -l | tr -d ' ')"
    if [ "$line_count" -le 1 ]; then
      echo "expected multi-line indented JSON, got $line_count line(s)" >&2
      echo "output: $output" >&2
      return 1
    fi
  fi
}

@test "--output json-pretty for list contains indentation" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json-pretty
  assert_success
  is_valid_json
  # Verify indentation is present when results exist (two-space indent per output.go)
  if [ "$output" != "[]" ]; then
    assert_output_contains "  "
  fi
}

# --- JSON Output (Detail / Get) ---

@test "--output json for get produces valid JSON object" {
  run "$RECURLY_BINARY" accounts get "code-test123" --output json
  assert_success
  is_valid_json
  # Detail view should be an object, not an array
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  if [ "$obj_type" != "object" ]; then
    echo "expected JSON object, got $obj_type" >&2
    return 1
  fi
}

@test "--output json for get contains expected account fields" {
  run "$RECURLY_BINARY" accounts get "code-test123" --output json
  assert_success
  is_valid_json
  # Verify expected fields exist
  local has_fields
  has_fields="$(echo "$output" | jq 'has("code") and has("email") and has("state")')"
  if [ "$has_fields" != "true" ]; then
    echo "expected JSON to have code, email, and state fields" >&2
    echo "output: $output" >&2
    return 1
  fi
}

# --- JSON Pretty Output (Detail / Get) ---

@test "--output json-pretty for get produces indented valid JSON" {
  run "$RECURLY_BINARY" accounts get "code-test123" --output json-pretty
  assert_success
  is_valid_json
  # Pretty JSON should have multiple lines
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  if [ "$line_count" -le 1 ]; then
    echo "expected multi-line indented JSON, got $line_count line(s)" >&2
    echo "output: $output" >&2
    return 1
  fi
}

# --- Default Table Output (Detail / Get) ---

@test "default output for get command shows key-value pairs" {
  run "$RECURLY_BINARY" accounts get "code-test123"
  assert_success
  # Detail view uses a "Field" / "Value" table
  assert_output_contains "Field"
  assert_output_contains "Value"
  assert_output_contains "Code"
  assert_output_contains "Email"
  assert_output_contains "State"
}

# --- Invalid Output Format ---

@test "invalid --output format produces an error" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output yaml
  assert_failure
  assert_output_contains "invalid output format"
}
