package cmd

import (
	"fmt"
	"os"
	"strings"

	huh "charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// stdinIsTerminal reports whether stdin is a terminal (not piped/scripted).
// Declared as a variable for testability.
var stdinIsTerminal = func() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

// flagOptionsKey is the annotation key for storing enum options on a flag.
const flagOptionsKey = "interactive_options"

// setFlagOptions annotates a flag with valid options for interactive select prompts.
func setFlagOptions(cmd *cobra.Command, flagName string, options []string) {
	f := cmd.Flags().Lookup(flagName)
	if f == nil {
		return
	}
	if f.Annotations == nil {
		f.Annotations = make(map[string][]string)
	}
	f.Annotations[flagOptionsKey] = options
}

// promptForFlag prompts the user for a single missing required flag value.
// Declared as a variable for testability.
var promptForFlag = defaultPromptForFlag

func defaultPromptForFlag(f *pflag.Flag) (string, error) {
	switch f.Value.Type() {
	case "bool":
		var value bool
		err := huh.NewConfirm().
			Title(f.Usage).
			Value(&value).
			Run()
		if err != nil {
			return "", err
		}
		if value {
			return "true", nil
		}
		return "false", nil

	default:
		// Check for enum options
		if opts, ok := f.Annotations[flagOptionsKey]; ok && len(opts) > 0 {
			var value string
			huhOpts := make([]huh.Option[string], len(opts))
			for i, o := range opts {
				huhOpts[i] = huh.NewOption(o, o)
			}
			err := huh.NewSelect[string]().
				Title(f.Usage).
				Options(huhOpts...).
				Value(&value).
				Run()
			return value, err
		}

		// Default: text input
		var value string
		input := huh.NewInput().
			Title(f.Usage).
			Value(&value)

		if strings.HasSuffix(f.Value.Type(), "Slice") {
			input = input.Description("Enter comma-separated values")
		}

		err := input.Run()
		return value, err
	}
}

// withInteractive wraps a command to prompt for missing required fields
// interactively when stdin is a TTY. When stdin is not a TTY or --no-input
// is set, the normal "required flag" error is shown instead.
func withInteractive(cmd *cobra.Command) *cobra.Command {
	var noInput bool
	cmd.Flags().BoolVar(&noInput, "no-input", false, "Disable interactive prompts for missing required fields")

	// Collect required flag names from Cobra annotations, then remove them
	// so Cobra doesn't validate before our wrapper runs.
	var requiredFlags []string
	cmd.Flags().VisitAll(func(f *pflag.Flag) {
		if anns, ok := f.Annotations[cobra.BashCompOneRequiredFlag]; ok && len(anns) > 0 {
			requiredFlags = append(requiredFlags, f.Name)
			delete(f.Annotations, cobra.BashCompOneRequiredFlag)
		}
	})

	if len(requiredFlags) == 0 {
		return cmd
	}

	origRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		// Find missing required flags
		var missing []*pflag.Flag
		for _, name := range requiredFlags {
			if !cmd.Flags().Changed(name) {
				if f := cmd.Flags().Lookup(name); f != nil {
					missing = append(missing, f)
				}
			}
		}

		if len(missing) > 0 {
			if noInput || !stdinIsTerminal() {
				names := make([]string, 0, len(missing))
				for _, f := range missing {
					names = append(names, fmt.Sprintf(`"--%s"`, f.Name))
				}
				return fmt.Errorf("required flag(s) %s not set", strings.Join(names, ", "))
			}

			for _, f := range missing {
				value, err := promptForFlag(f)
				if err != nil {
					return fmt.Errorf("prompt for --%s: %w", f.Name, err)
				}
				if err := cmd.Flags().Set(f.Name, value); err != nil {
					return fmt.Errorf("setting --%s: %w", f.Name, err)
				}
			}
		}

		return origRunE(cmd, args)
	}

	return cmd
}
