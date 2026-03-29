package cmd

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// newInteractiveTestCmd creates a simple command for interactive wrapper tests.
func newInteractiveTestCmd(runFn func(cmd *cobra.Command, args []string) error) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "test",
		SilenceErrors: true,
		SilenceUsage:  true,
		RunE:          runFn,
	}
	return cmd
}

func TestWithInteractive_MissingRequiredFlags_TTY(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	origPrompt := promptForFlag
	defer func() {
		stdinIsTerminal = origIsTerminal
		promptForFlag = origPrompt
	}()

	stdinIsTerminal = func() bool { return true }

	var prompted []string
	promptForFlag = func(f *pflag.Flag) (string, error) {
		prompted = append(prompted, f.Name)
		return "test-value", nil
	}

	var ranCmd bool
	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error {
		ranCmd = true
		return nil
	})
	cmd.Flags().String("name", "", "Name (required)")
	_ = cmd.MarkFlagRequired("name")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{})
	err := wrapped.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ranCmd {
		t.Fatal("expected command to run")
	}
	if len(prompted) != 1 || prompted[0] != "name" {
		t.Errorf("expected prompt for 'name', got %v", prompted)
	}
}

func TestWithInteractive_MissingRequiredFlags_NoTTY(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	defer func() { stdinIsTerminal = origIsTerminal }()

	stdinIsTerminal = func() bool { return false }

	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error {
		return nil
	})
	cmd.Flags().String("name", "", "Name (required)")
	_ = cmd.MarkFlagRequired("name")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{})
	err := wrapped.Execute()
	if err == nil {
		t.Fatal("expected error for missing required flag")
	}
	if !strings.Contains(err.Error(), `"--name"`) {
		t.Errorf("expected error about --name, got: %s", err.Error())
	}
}

func TestWithInteractive_NoInputFlag(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	defer func() { stdinIsTerminal = origIsTerminal }()

	stdinIsTerminal = func() bool { return true }

	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error {
		return nil
	})
	cmd.Flags().String("name", "", "Name (required)")
	_ = cmd.MarkFlagRequired("name")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{"--no-input"})
	err := wrapped.Execute()
	if err == nil {
		t.Fatal("expected error for missing required flag with --no-input")
	}
	if !strings.Contains(err.Error(), `"--name"`) {
		t.Errorf("expected error about --name, got: %s", err.Error())
	}
}

func TestWithInteractive_FlagProvidedOverridesPrompt(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	origPrompt := promptForFlag
	defer func() {
		stdinIsTerminal = origIsTerminal
		promptForFlag = origPrompt
	}()

	stdinIsTerminal = func() bool { return true }

	var prompted []string
	promptForFlag = func(f *pflag.Flag) (string, error) {
		prompted = append(prompted, f.Name)
		return "prompted-value", nil
	}

	var gotName, gotCode string
	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error {
		gotName, _ = cmd.Flags().GetString("name")
		gotCode, _ = cmd.Flags().GetString("code")
		return nil
	})
	cmd.Flags().String("name", "", "Name (required)")
	cmd.Flags().String("code", "", "Code (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("code")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{"--name", "cli-value"})
	err := wrapped.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(prompted) != 1 || prompted[0] != "code" {
		t.Errorf("expected prompt only for 'code', got %v", prompted)
	}
	if gotName != "cli-value" {
		t.Errorf("expected name 'cli-value', got %q", gotName)
	}
	if gotCode != "prompted-value" {
		t.Errorf("expected code 'prompted-value', got %q", gotCode)
	}
}

func TestWithInteractive_AllFlagsProvided_NoPrompt(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	origPrompt := promptForFlag
	defer func() {
		stdinIsTerminal = origIsTerminal
		promptForFlag = origPrompt
	}()

	stdinIsTerminal = func() bool { return true }

	var prompted []string
	promptForFlag = func(f *pflag.Flag) (string, error) {
		prompted = append(prompted, f.Name)
		return "prompted-value", nil
	}

	var ranCmd bool
	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error {
		ranCmd = true
		return nil
	})
	cmd.Flags().String("name", "", "Name (required)")
	_ = cmd.MarkFlagRequired("name")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{"--name", "provided"})
	err := wrapped.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ranCmd {
		t.Fatal("expected command to run")
	}
	if len(prompted) != 0 {
		t.Errorf("expected no prompts, got %v", prompted)
	}
}

func TestWithInteractive_BoolFlag(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	origPrompt := promptForFlag
	defer func() {
		stdinIsTerminal = origIsTerminal
		promptForFlag = origPrompt
	}()

	stdinIsTerminal = func() bool { return true }

	var promptedTypes []string
	promptForFlag = func(f *pflag.Flag) (string, error) {
		promptedTypes = append(promptedTypes, f.Value.Type())
		return "true", nil
	}

	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error { return nil })
	cmd.Flags().Bool("active", false, "Is active (required)")
	_ = cmd.MarkFlagRequired("active")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{})
	err := wrapped.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(promptedTypes) != 1 || promptedTypes[0] != "bool" {
		t.Errorf("expected bool type prompt, got %v", promptedTypes)
	}
}

func TestWithInteractive_EnumFlag(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	origPrompt := promptForFlag
	defer func() {
		stdinIsTerminal = origIsTerminal
		promptForFlag = origPrompt
	}()

	stdinIsTerminal = func() bool { return true }

	var sawOptions bool
	promptForFlag = func(f *pflag.Flag) (string, error) {
		if opts, ok := f.Annotations[flagOptionsKey]; ok && len(opts) > 0 {
			sawOptions = true
		}
		return "day", nil
	}

	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error { return nil })
	cmd.Flags().String("unit", "", "Time unit (required)")
	_ = cmd.MarkFlagRequired("unit")
	setFlagOptions(cmd, "unit", []string{"day", "week", "month"})

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{})
	err := wrapped.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !sawOptions {
		t.Error("expected enum options annotation to be present")
	}
}

func TestWithInteractive_MultipleRequiredFlags_ErrorMessage(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	defer func() { stdinIsTerminal = origIsTerminal }()

	stdinIsTerminal = func() bool { return false }

	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error { return nil })
	cmd.Flags().String("name", "", "Name (required)")
	cmd.Flags().String("code", "", "Code (required)")
	_ = cmd.MarkFlagRequired("name")
	_ = cmd.MarkFlagRequired("code")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{})
	err := wrapped.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	errMsg := err.Error()
	if !strings.Contains(errMsg, `"--name"`) || !strings.Contains(errMsg, `"--code"`) {
		t.Errorf("expected error mentioning both flags, got: %s", errMsg)
	}
}

func TestWithInteractive_NoRequiredFlags(t *testing.T) {
	var ranCmd bool
	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error {
		ranCmd = true
		return nil
	})
	cmd.Flags().String("name", "", "Name")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{})
	err := wrapped.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ranCmd {
		t.Fatal("expected command to run")
	}
}

func TestWithInteractive_HelpDisplay(t *testing.T) {
	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error { return nil })
	cmd.Flags().String("name", "", "Name (required)")
	_ = cmd.MarkFlagRequired("name")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{"--help"})

	err := wrapped.Execute()
	if err != nil {
		t.Fatalf("help should not error: %v", err)
	}
}

func TestWithInteractive_SliceFlag(t *testing.T) {
	origIsTerminal := stdinIsTerminal
	origPrompt := promptForFlag
	defer func() {
		stdinIsTerminal = origIsTerminal
		promptForFlag = origPrompt
	}()

	stdinIsTerminal = func() bool { return true }

	var promptedType string
	promptForFlag = func(f *pflag.Flag) (string, error) {
		promptedType = f.Value.Type()
		return "USD,EUR", nil
	}

	var gotCurrencies []string
	cmd := newInteractiveTestCmd(func(cmd *cobra.Command, args []string) error {
		gotCurrencies, _ = cmd.Flags().GetStringSlice("currency")
		return nil
	})
	cmd.Flags().StringSlice("currency", nil, "Currency codes (required)")
	_ = cmd.MarkFlagRequired("currency")

	wrapped := withInteractive(cmd)
	wrapped.SetArgs([]string{})
	err := wrapped.Execute()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if promptedType != "stringSlice" {
		t.Errorf("expected stringSlice type, got %q", promptedType)
	}
	if len(gotCurrencies) != 2 || gotCurrencies[0] != "USD" || gotCurrencies[1] != "EUR" {
		t.Errorf("expected [USD EUR], got %v", gotCurrencies)
	}
}

func TestSetFlagOptions(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Flags().String("unit", "", "Time unit")

	setFlagOptions(cmd, "unit", []string{"day", "week", "month"})

	f := cmd.Flags().Lookup("unit")
	opts, ok := f.Annotations[flagOptionsKey]
	if !ok {
		t.Fatal("expected annotation to be set")
	}
	if len(opts) != 3 || opts[0] != "day" || opts[1] != "week" || opts[2] != "month" {
		t.Errorf("unexpected options: %v", opts)
	}
}

func TestSetFlagOptions_NonexistentFlag(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	// Should not panic
	setFlagOptions(cmd, "nonexistent", []string{"a", "b"})
}
