package cmd

import (
	"strings"
	"testing"
)

func TestAccountsList_ShowsInHelp(t *testing.T) {
	out, _, err := executeCommand("accounts", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "list") {
		t.Error("expected accounts help to show 'list' subcommand")
	}
}

func TestAccountsListHelp_ShowsFlags(t *testing.T) {
	out, _, err := executeCommand("accounts", "list", "--help")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, flag := range []string{"--limit", "--all", "--order", "--sort", "--email", "--subscriber", "--past-due", "--begin-time", "--end-time"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected help output to contain flag %q", flag)
		}
	}
}

func TestAccountsList_NoAPIKey_ReturnsError(t *testing.T) {
	// Without an API key configured, the command should fail
	t.Setenv("RECURLY_API_KEY", "")
	_, stderr, err := executeCommand("accounts", "list")
	if err == nil {
		t.Fatal("expected error when no API key is configured")
	}
	if !strings.Contains(stderr, "API key not configured") {
		t.Errorf("expected 'API key not configured' error, got %q", stderr)
	}
}

func TestAccountsList_InvalidBeginTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand("accounts", "list", "--begin-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid begin-time")
	}
	if !strings.Contains(stderr, "invalid --begin-time") {
		t.Errorf("expected 'invalid --begin-time' error, got %q", stderr)
	}
}

func TestAccountsList_InvalidEndTime_ReturnsError(t *testing.T) {
	t.Setenv("RECURLY_API_KEY", "test-key")
	_, stderr, err := executeCommand("accounts", "list", "--end-time", "not-a-date")
	if err == nil {
		t.Fatal("expected error for invalid end-time")
	}
	if !strings.Contains(stderr, "invalid --end-time") {
		t.Errorf("expected 'invalid --end-time' error, got %q", stderr)
	}
}
