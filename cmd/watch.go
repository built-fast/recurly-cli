package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/spf13/cobra"
)

// stdoutIsPiped reports whether stdout is a pipe (not a terminal).
// Declared as a variable for testability.
var stdoutIsPiped = func() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) == 0
}

// withWatch wraps a get command to add --watch support.
// When --watch is specified, the command re-runs on a polling interval,
// refreshing the display in-place for terminal output or emitting
// newline-delimited JSON when stdout is piped.
func withWatch(cmd *cobra.Command) *cobra.Command {
	cmd.Flags().String("watch", "", "Watch resource with polling interval (e.g., 5s, 1m, 30s; default 5s)")
	cmd.Flags().Lookup("watch").NoOptDefVal = "5s"

	origRunE := cmd.RunE
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if !cmd.Flags().Changed("watch") {
			return origRunE(cmd, args)
		}

		watchStr, _ := cmd.Flags().GetString("watch")
		interval, err := parseWatchInterval(watchStr)
		if err != nil {
			return err
		}

		cfg := output.FromContext(cmd.Context())
		piped := stdoutIsPiped()

		if isJSONFormat(cfg.Format) && !piped {
			return fmt.Errorf("--watch is incompatible with --output %s in interactive mode; pipe output to use JSON with watch", cfg.Format)
		}

		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()

		watchLoop(ctx, cmd, args, origRunE, interval, piped)
		return nil
	}

	return cmd
}

// watchLoop polls the command on the given interval until the context is canceled.
func watchLoop(ctx context.Context, cmd *cobra.Command, args []string, runE func(*cobra.Command, []string) error, interval time.Duration, piped bool) {
	runWatchIteration(cmd, args, runE, piped, interval)

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			if !piped {
				_, _ = fmt.Fprintln(cmd.ErrOrStderr(), "\nWatch stopped.")
			}
			return
		case <-ticker.C:
			runWatchIteration(cmd, args, runE, piped, interval)
		}
	}
}

// runWatchIteration executes a single poll: clears the screen (for terminals),
// runs the command, and prints a timestamp footer.
func runWatchIteration(cmd *cobra.Command, args []string, runE func(*cobra.Command, []string) error, piped bool, interval time.Duration) {
	if !piped {
		_, _ = fmt.Fprint(cmd.OutOrStdout(), "\033[2J\033[H")
	}

	err := runE(cmd, args)
	if err != nil {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", err)
	}

	if !piped {
		_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "\nLast updated: %s (every %s, Ctrl+C to stop)\n",
			time.Now().Format("15:04:05"), interval)
	}
}

// parseWatchInterval parses a duration string and validates the minimum interval.
func parseWatchInterval(s string) (time.Duration, error) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 0, fmt.Errorf("invalid watch interval %q: use duration strings like 5s, 1m, 30s", s)
	}
	if d < time.Second {
		return 0, fmt.Errorf("watch interval must be at least 1s")
	}
	return d, nil
}

func isJSONFormat(format string) bool {
	return format == "json" || format == "json-pretty"
}
