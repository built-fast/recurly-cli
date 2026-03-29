package cmd

import (
	"strings"
	"testing"
)

func TestCompletionBash_OutputsScript(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "completion", "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bash") || !strings.Contains(out, "complete") {
		t.Error("expected bash completion script output")
	}
}

func TestCompletionZsh_OutputsScript(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "completion", "zsh")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "zsh") || !strings.Contains(out, "compdef") {
		t.Error("expected zsh completion script output")
	}
}

func TestCompletionFish_OutputsScript(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "completion", "fish")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "fish") || !strings.Contains(out, "complete") {
		t.Error("expected fish completion script output")
	}
}

func TestCompletionPowershell_OutputsScript(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "completion", "powershell")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "PowerShell") || !strings.Contains(out, "Register-ArgumentCompleter") {
		t.Error("expected PowerShell completion script output")
	}
}

func TestCompletionBash_IncludesCompletionFunction(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "completion", "bash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "__recurly") {
		t.Error("expected bash completion to include __recurly completion function")
	}
}

func TestCompletionNoArgs_ShowsHelp(t *testing.T) {
	t.Parallel()
	out, _, err := executeCommand(nil, "completion")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(out, "bash") || !strings.Contains(out, "zsh") {
		t.Error("expected completion help to list shell subcommands")
	}
}

func TestCompletion_HasUsageExamples(t *testing.T) {
	t.Parallel()
	shells := []string{"bash", "zsh", "fish", "powershell"}
	for _, shell := range shells {
		out, _, err := executeCommand(nil, "completion", shell, "--help")
		if err != nil {
			t.Fatalf("unexpected error for %s: %v", shell, err)
		}
		if !strings.Contains(out, "Examples:") {
			t.Errorf("expected %s completion help to contain usage examples", shell)
		}
	}
}
