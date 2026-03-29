#!/usr/bin/env bats

load "test_helper"

# =============================================================================
# Rate Limit Retry Tests
# Prism does not support returning 429 responses (the Recurly OpenAPI spec
# does not define 429 as a response code), so these tests are skipped.
# Rate limit retry logic is covered by unit tests in internal/client/retry_test.go.
# =============================================================================

@test "rate limit retry on 429 response" {
  skip "Prism cannot mock 429 responses (not defined in OpenAPI spec)"
}
