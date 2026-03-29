#!/usr/bin/env bats

load "test_helper"

# =============================================================================
# --from-file / -F Flag Tests
# Tests verify JSON and YAML file input on the accounts create command.
# =============================================================================

# --- JSON File Input ---

@test "accounts create --from-file with JSON file returns success" {
  local json_file="$TEST_TEMP_DIR/account.json"
  cat > "$json_file" <<'EOF'
{
  "code": "e2e-fromfile-json",
  "email": "fromfile@example.com",
  "first_name": "FromFile",
  "last_name": "JSON"
}
EOF
  run "$RECURLY_BINARY" accounts create --from-file "$json_file"
  assert_success
}

@test "accounts create --from-file JSON with --output json returns valid JSON" {
  local json_file="$TEST_TEMP_DIR/account.json"
  cat > "$json_file" <<'EOF'
{
  "code": "e2e-fromfile-json2",
  "email": "fromfile2@example.com"
}
EOF
  run "$RECURLY_BINARY" accounts create --from-file "$json_file" --output json
  assert_success
  is_valid_json
}

# --- YAML File Input ---

@test "accounts create --from-file with YAML file returns success" {
  local yaml_file="$TEST_TEMP_DIR/account.yaml"
  cat > "$yaml_file" <<'EOF'
code: e2e-fromfile-yaml
email: fromfile-yaml@example.com
first_name: FromFile
last_name: YAML
EOF
  run "$RECURLY_BINARY" accounts create --from-file "$yaml_file"
  assert_success
}

@test "accounts create --from-file with .yml extension returns success" {
  local yml_file="$TEST_TEMP_DIR/account.yml"
  cat > "$yml_file" <<'EOF'
code: e2e-fromfile-yml
email: fromfile-yml@example.com
EOF
  run "$RECURLY_BINARY" accounts create --from-file "$yml_file"
  assert_success
}

@test "accounts create --from-file YAML with --output json returns valid JSON" {
  local yaml_file="$TEST_TEMP_DIR/account.yaml"
  cat > "$yaml_file" <<'EOF'
code: e2e-fromfile-yaml2
email: fromfile-yaml2@example.com
EOF
  run "$RECURLY_BINARY" accounts create --from-file "$yaml_file" --output json
  assert_success
  is_valid_json
}

# --- CLI Flag Override ---

@test "CLI flag overrides file value" {
  local json_file="$TEST_TEMP_DIR/account.json"
  cat > "$json_file" <<'EOF'
{
  "code": "file-code",
  "email": "file@example.com"
}
EOF
  run "$RECURLY_BINARY" accounts create --from-file "$json_file" --code "cli-override-code" --output json
  assert_success
  is_valid_json
}

# --- Short Flag ---

@test "accounts create -F (short flag) works" {
  local json_file="$TEST_TEMP_DIR/account.json"
  cat > "$json_file" <<'EOF'
{
  "code": "e2e-short-flag"
}
EOF
  run "$RECURLY_BINARY" accounts create -F "$json_file"
  assert_success
}

# --- Error Cases ---

@test "--from-file with nonexistent file fails" {
  run "$RECURLY_BINARY" accounts create --from-file "/nonexistent/path.json"
  assert_failure
  assert_output_contains "error reading file"
}

@test "--from-file with unsupported extension fails" {
  local txt_file="$TEST_TEMP_DIR/account.txt"
  echo '{"code": "test"}' > "$txt_file"
  run "$RECURLY_BINARY" accounts create --from-file "$txt_file"
  assert_failure
  assert_output_contains "unsupported file extension"
}

@test "--from-file with unknown key fails" {
  local json_file="$TEST_TEMP_DIR/account.json"
  cat > "$json_file" <<'EOF'
{
  "code": "test",
  "nonexistent_field": "value"
}
EOF
  run "$RECURLY_BINARY" accounts create --from-file "$json_file"
  assert_failure
  assert_output_contains "unknown key"
}

@test "--from-file with invalid JSON fails" {
  local json_file="$TEST_TEMP_DIR/bad.json"
  echo 'not valid json{{{' > "$json_file"
  run "$RECURLY_BINARY" accounts create --from-file "$json_file"
  assert_failure
  assert_output_contains "invalid JSON"
}
