package cmd

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/built-fast/recurly-cli/internal/client"
)

func executeCommand(args ...string) (string, string, error) {
	return executeCommandWithStdin(nil, args...)
}

// testApp, when non-nil, is injected into the root command's context by
// executeCommandWithStdin so tests can supply mock API factories via App
// instead of overriding package-level vars.
var testApp *App

func executeCommandWithStdin(stdin *bytes.Buffer, args ...string) (string, string, error) {
	cmd := NewRootCmd()
	if testApp != nil {
		cmd.SetContext(NewAppContext(cmd.Context(), testApp))
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	if stdin != nil {
		cmd.SetIn(stdin)
	}
	cmd.SetArgs(args)

	err := cmd.Execute()
	if err != nil {
		fmt.Fprintln(stderr, client.FormatError(err))
	}
	return stdout.String(), stderr.String(), err
}

func TestRootNoArgs_ShowsHelp(t *testing.T) {
	out, _, err := executeCommand()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "Recurly CLI") {
		t.Error("expected help output to mention 'Recurly CLI'")
	}
	if !strings.Contains(out, "Available Commands") {
		t.Error("expected help output to list available commands")
	}
	for _, flag := range []string{"--api-key", "--region", "--output"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain global flag %q", flag)
		}
	}
}

func TestRootHelp_ShowsHelp(t *testing.T) {
	out, _, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--api-key", "--region", "--output"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain global flag %q", flag)
		}
	}
}

func TestRootVersion_PrintsVersionString(t *testing.T) {
	out, _, err := executeCommand("--version")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	expected := "recurly-cli dev\n"
	if out != expected {
		t.Errorf("expected %q, got %q", expected, out)
	}
}

func TestRootUnknownCommand_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand("nonexistent-command")
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
	if !strings.Contains(stderr, "unknown command") {
		t.Errorf("expected stderr to contain 'unknown command', got %q", stderr)
	}
}

func TestRootInvalidRegion_ReturnsError(t *testing.T) {
	_, stderr, err := executeCommand("--region", "asia", "configure")
	if err == nil {
		t.Fatal("expected error for invalid region")
	}
	if !strings.Contains(stderr, "invalid region") {
		t.Errorf("expected stderr to contain 'invalid region', got %q", stderr)
	}
	if !strings.Contains(stderr, "us, eu") {
		t.Errorf("expected stderr to list valid options, got %q", stderr)
	}
}

func TestRootValidRegion_CaseInsensitive(t *testing.T) {
	// Passing --region=EU should not cause an error (help output for configure)
	out, _, err := executeCommand("--region", "EU", "--help")
	if err != nil {
		t.Fatalf("unexpected error with --region EU: %v", err)
	}
	if !strings.Contains(out, "Recurly CLI") {
		t.Error("expected help output")
	}
}

func TestRootJQFlag_InvalidExpression(t *testing.T) {
	_, stderr, err := executeCommand("--jq", "invalid[[[", "configure")
	if err == nil {
		t.Fatal("expected error for invalid jq expression")
	}
	if !strings.Contains(stderr, "invalid jq expression") {
		t.Errorf("expected stderr to contain 'invalid jq expression', got %q", stderr)
	}
}

func TestRootJQFlag_MutuallyExclusiveWithTable(t *testing.T) {
	_, stderr, err := executeCommand("--jq", ".name", "--output", "table", "configure")
	if err == nil {
		t.Fatal("expected error for --jq with --output table")
	}
	if !strings.Contains(stderr, "mutually exclusive") {
		t.Errorf("expected stderr to contain 'mutually exclusive', got %q", stderr)
	}
}

func TestRootJQFlag_AllowedWithJSON(t *testing.T) {
	// --jq with --output json should not produce a mutual exclusivity error
	_, stderr, err := executeCommand("--jq", ".", "--output", "json", "configure")
	if err != nil && strings.Contains(stderr, "mutually exclusive") {
		t.Error("--jq with --output json should not be mutually exclusive")
	}
}

func TestRootJQFlag_AllowedWithJSONPretty(t *testing.T) {
	// --jq with --output json-pretty should not produce a mutual exclusivity error
	_, stderr, err := executeCommand("--jq", ".", "--output", "json-pretty", "configure")
	if err != nil && strings.Contains(stderr, "mutually exclusive") {
		t.Error("--jq with --output json-pretty should not be mutually exclusive")
	}
}

func TestRootJQFlag_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "--jq") {
		t.Error("expected help output to contain --jq flag")
	}
}
