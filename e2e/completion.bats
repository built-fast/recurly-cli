#!/usr/bin/env bats

load "test_helper"

# =============================================================================
# Shell Completion Commands
# =============================================================================

@test "completion bash produces valid bash completion script" {
  run "$RECURLY_BINARY" completion bash
  assert_success
  # Bash completions contain specific shell functions
  assert_output_contains "complete"
  assert_output_contains "recurly"
}

@test "completion zsh produces valid zsh completion script" {
  run "$RECURLY_BINARY" completion zsh
  assert_success
  # Zsh completions use compdef
  assert_output_contains "compdef"
  assert_output_contains "recurly"
}

@test "completion fish produces valid fish completion script" {
  run "$RECURLY_BINARY" completion fish
  assert_success
  # Fish completions use the complete command
  assert_output_contains "complete"
  assert_output_contains "recurly"
}

@test "completion powershell produces valid powershell completion script" {
  run "$RECURLY_BINARY" completion powershell
  assert_success
  # PowerShell completions use Register-ArgumentCompleter
  assert_output_contains "Register-ArgumentCompleter"
  assert_output_contains "recurly"
}

@test "completion without subcommand shows help" {
  run "$RECURLY_BINARY" completion --help
  assert_success
  assert_output_contains "bash"
  assert_output_contains "zsh"
  assert_output_contains "fish"
  assert_output_contains "powershell"
}

@test "completion bash output is non-trivial" {
  run "$RECURLY_BINARY" completion bash
  assert_success
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 10 ]
}

@test "completion zsh output is non-trivial" {
  run "$RECURLY_BINARY" completion zsh
  assert_success
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 10 ]
}

@test "completion fish output is non-trivial" {
  run "$RECURLY_BINARY" completion fish
  assert_success
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 10 ]
}

@test "completion powershell output is non-trivial" {
  run "$RECURLY_BINARY" completion powershell
  assert_success
  local line_count
  line_count="$(echo "$output" | wc -l | tr -d ' ')"
  [ "$line_count" -gt 10 ]
}
