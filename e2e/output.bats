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

@test "--output json for list produces a list envelope" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json
  assert_success
  is_valid_json
  # Verify it's an object with list envelope fields
  local obj_type
  obj_type="$(echo "$output" | jq -r 'type')"
  if [ "$obj_type" != "object" ]; then
    echo "expected JSON object (list envelope), got $obj_type" >&2
    return 1
  fi
  local object_field
  object_field="$(echo "$output" | jq -r '.object')"
  if [ "$object_field" != "list" ]; then
    echo "expected .object=list, got $object_field" >&2
    return 1
  fi
  # Verify has_more is a boolean
  local has_more_type
  has_more_type="$(echo "$output" | jq -r '.has_more | type')"
  if [ "$has_more_type" != "boolean" ]; then
    echo "expected .has_more to be boolean, got $has_more_type" >&2
    return 1
  fi
  # Verify data is an array
  local data_type
  data_type="$(echo "$output" | jq -r '.data | type')"
  if [ "$data_type" != "array" ]; then
    echo "expected .data to be array, got $data_type" >&2
    return 1
  fi
}

@test "--output json for list contains expected top-level fields" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json
  assert_success
  is_valid_json
  # Verify expected account fields exist in the first element of .data
  local has_code
  has_code="$(echo "$output" | jq '.data[0] | has("code")')"
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
  # Envelope always has multiple lines in pretty mode
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  if [ "$line_count" -le 1 ]; then
    echo "expected multi-line indented JSON, got $line_count line(s)" >&2
    echo "output: $output" >&2
    return 1
  fi
}

@test "--output json-pretty for list contains indentation" {
  run "$RECURLY_BINARY" accounts list --limit 1 --output json-pretty
  assert_success
  is_valid_json
  # Verify indentation is present (two-space indent per output.go)
  assert_output_contains "  "
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

# --- jq Filtering (List) ---

@test "--jq on list receives full envelope with .object field" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.object'
  assert_success
  if [ "$output" != "list" ]; then
    echo "expected 'list', got '$output'" >&2
    return 1
  fi
}

@test "--jq on list can access .has_more boolean" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.has_more'
  assert_success
  # has_more should be true or false
  if [ "$output" != "true" ] && [ "$output" != "false" ]; then
    echo "expected true or false, got '$output'" >&2
    return 1
  fi
}

@test "--jq on list can access .data array" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data | type'
  assert_success
  if [ "$output" != "array" ]; then
    echo "expected 'array', got '$output'" >&2
    return 1
  fi
}

@test "--jq on list extracts field from data items" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data[0].code'
  assert_success
  # Should produce a non-empty string (the account code)
  if [ -z "$output" ]; then
    echo "expected non-empty output for .data[0].code" >&2
    return 1
  fi
}

@test "--jq on list with .data[] produces multiple lines" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data[].code'
  assert_success
  # At least one line of output
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  if [ "$line_count" -lt 1 ]; then
    echo "expected at least 1 line, got $line_count" >&2
    return 1
  fi
}

@test "--jq string result prints raw (no quotes)" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.object'
  assert_success
  # Should not have surrounding quotes
  if [[ "$output" == \"*\" ]]; then
    echo "expected raw string without quotes, got '$output'" >&2
    return 1
  fi
}

@test "--jq null result prints literal null" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.nonexistent_field'
  assert_success
  if [ "$output" != "null" ]; then
    echo "expected 'null', got '$output'" >&2
    return 1
  fi
}

@test "--jq numeric result prints number" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data | length'
  assert_success
  # Should be a number (at least 1 since we got results)
  if ! [[ "$output" =~ ^[0-9]+$ ]]; then
    echo "expected numeric output, got '$output'" >&2
    return 1
  fi
}

@test "--jq object result with --output json is compact" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data[0] | {code: .code}' --output json
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  if [ "$line_count" -ne 1 ]; then
    echo "expected compact JSON on 1 line, got $line_count" >&2
    return 1
  fi
}

@test "--jq object result with --output json-pretty is indented" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data[0] | {code: .code}' --output json-pretty
  assert_success
  is_valid_json
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  if [ "$line_count" -le 1 ]; then
    echo "expected indented JSON (multi-line), got $line_count line(s)" >&2
    return 1
  fi
}

@test "--jq empty result produces no output" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data[] | select(.code == "IMPOSSIBLE_CODE_VALUE_XYZ")'
  assert_success
  if [ -n "$output" ]; then
    echo "expected empty output, got '$output'" >&2
    return 1
  fi
}

# --- jq Filtering (Detail / Get) ---

@test "--jq on get receives bare resource object" {
  run "$RECURLY_BINARY" accounts get "code-test123" --jq '.code'
  assert_success
  # Should produce a non-empty string (the account code)
  if [ -z "$output" ]; then
    echo "expected non-empty output for .code" >&2
    return 1
  fi
}

@test "--jq on get extracts .email as raw string" {
  run "$RECURLY_BINARY" accounts get "code-test123" --jq '.email'
  assert_success
  # Should produce a non-empty raw string (no surrounding quotes)
  if [ -z "$output" ]; then
    echo "expected non-empty output for .email" >&2
    return 1
  fi
  if [[ "$output" == \"*\" ]]; then
    echo "expected raw string without quotes, got '$output'" >&2
    return 1
  fi
}

@test "--jq on get does not have envelope fields" {
  run "$RECURLY_BINARY" accounts get "code-test123" --jq '.object'
  assert_success
  # For a bare resource, .object is typically "account" or similar —
  # but definitely not "list" (which would indicate envelope wrapping)
  if [ "$output" = "list" ]; then
    echo "expected bare resource (not list envelope)" >&2
    return 1
  fi
}

# --- jq Advanced Filtering ---

@test "--jq select filter counts non-null states" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '[.data[] | select(.state != null)] | length'
  assert_success
  # Should be a number (accounts with non-null state)
  if ! [[ "$output" =~ ^[0-9]+$ ]]; then
    echo "expected numeric output, got '$output'" >&2
    return 1
  fi
}

@test "--jq @csv format produces CSV output" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data[] | [.code, .email] | @csv'
  assert_success
  # CSV output should contain comma-separated quoted values
  if [ -z "$output" ]; then
    echo "expected non-empty CSV output" >&2
    return 1
  fi
  assert_output_contains ","
}

@test "--jq @base64 encodes field value" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.data[0].code | @base64'
  assert_success
  # Base64 output should be non-empty and contain only valid base64 chars
  if [ -z "$output" ]; then
    echo "expected non-empty base64 output" >&2
    return 1
  fi
  if ! [[ "$output" =~ ^[A-Za-z0-9+/=]+$ ]]; then
    echo "expected valid base64, got '$output'" >&2
    return 1
  fi
}

# --- jq Error Handling ---

@test "--jq with invalid expression exits non-zero" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq 'invalid[['
  assert_failure
}

@test "--jq with --output table exits non-zero (mutual exclusion)" {
  run "$RECURLY_BINARY" accounts list --limit 1 --jq '.code' --output table
  assert_failure
  assert_output_contains "mutually exclusive"
}
