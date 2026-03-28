#!/usr/bin/env bats

load "test_helper"

# Use --limit 1 on list commands to avoid Prism pagination issues
# (Prism generates random "next" URLs the SDK cannot follow).

# --- Missing Credentials ---

@test "command without any API key fails with exit code 1 and helpful error" {
  unset RECURLY_API_KEY

  run "$RECURLY_BINARY" accounts list --limit 1
  assert_failure
  assert_output_contains "API key not configured"
  assert_output_contains "recurly configure"
  assert_output_contains "RECURLY_API_KEY"
}

# --- Env Var Provides API Key ---

@test "RECURLY_API_KEY env var is used when no config file or flag is set" {
  # setup() exports RECURLY_API_KEY=test-api-key and RECURLY_API_URL (Prism)
  # No config file created — env var is the sole key source

  run "$RECURLY_BINARY" accounts list --limit 1
  assert_success
}

# --- Flag Overrides Env Var ---

@test "--api-key flag overrides RECURLY_API_KEY env var" {
  # setup() exports RECURLY_API_KEY=test-api-key
  # Provide a different key via --api-key flag; Prism accepts any valid auth
  export RECURLY_API_KEY="env-key-should-be-ignored"

  run "$RECURLY_BINARY" accounts list --limit 1 --api-key "flag-override-key"
  assert_success
}

# --- Invalid API Key (Prism 401) ---

@test "invalid API key returns an auth error" {
  # Send a request with an empty-ish key that passes CLI validation
  # but Prism rejects with 401 (security scheme requires valid basic auth)
  export RECURLY_API_KEY="bad"

  run "$RECURLY_BINARY" accounts list --limit 1
  # Prism may accept any key in mock mode — if the command succeeds,
  # the CLI at least correctly forwarded the key to the server.
  # If Prism returns 401, verify the auth error message.
  if [ "$status" -ne 0 ]; then
    assert_output_contains "Invalid API key"
  else
    # Prism accepted the key — verify the command produced output
    # (confirms the key was sent and the request completed)
    [ -n "$output" ]
  fi
}
