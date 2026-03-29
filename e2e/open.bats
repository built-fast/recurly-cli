#!/usr/bin/env bats

load "test_helper"

# =============================================================================
# recurly open --url
# The --url flag prints the URL instead of opening a browser.
# A --site flag is required (no configure step in e2e).
# =============================================================================

@test "open --url with no args prints dashboard home URL" {
  run "$RECURLY_BINARY" open --url --site testsite
  assert_success
  assert_output_contains "https://testsite.recurly.com"
}

@test "open --url accounts prints accounts list URL" {
  run "$RECURLY_BINARY" open --url --site testsite accounts
  assert_success
  assert_output_contains "https://testsite.recurly.com/accounts"
}

@test "open --url accounts with identifier prints account URL" {
  run "$RECURLY_BINARY" open --url --site testsite accounts code-test123
  assert_success
  assert_output_contains "https://testsite.recurly.com/accounts/code-test123"
}

@test "open --url plans prints plans URL" {
  run "$RECURLY_BINARY" open --url --site testsite plans
  assert_success
  assert_output_contains "https://testsite.recurly.com/plans"
}

@test "open --url items prints items URL" {
  run "$RECURLY_BINARY" open --url --site testsite items
  assert_success
  assert_output_contains "https://testsite.recurly.com/items"
}

@test "open --url invoices prints invoices URL" {
  run "$RECURLY_BINARY" open --url --site testsite invoices
  assert_success
  assert_output_contains "https://testsite.recurly.com/invoices"
}

@test "open --url coupons prints coupons URL" {
  run "$RECURLY_BINARY" open --url --site testsite coupons
  assert_success
  assert_output_contains "https://testsite.recurly.com/coupons"
}

@test "open with invalid resource type fails" {
  run "$RECURLY_BINARY" open --url --site testsite invalid_resource
  assert_failure
  assert_output_contains "unrecognized resource type"
}

@test "open without --site and no config fails" {
  run "$RECURLY_BINARY" open --url accounts
  assert_failure
  assert_output_contains "site subdomain is required"
}

@test "open with too many arguments fails" {
  run "$RECURLY_BINARY" open --url --site testsite accounts code1 extra
  assert_failure
}

@test "open --help shows usage" {
  run "$RECURLY_BINARY" open --help
  assert_success
  assert_output_contains "Open a Recurly resource"
  assert_output_contains "--url"
  assert_output_contains "--site"
}
