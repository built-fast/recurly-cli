package cmd

import (
	"bufio"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

// confirm prompts the user for confirmation via cmd's stdin/stderr.
// It returns true if the user enters "y" or "yes" (case-insensitive),
// false otherwise. The message should be a complete prompt string
// (e.g. "Are you sure you want to deactivate this account? [y/N] ").
func confirm(cmd *cobra.Command, message string) (bool, error) {
	if _, err := fmt.Fprint(cmd.ErrOrStderr(), message); err != nil {
		return false, err
	}
	reader := bufio.NewReader(cmd.InOrStdin())
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return false, fmt.Errorf("reading confirmation: %w", err)
	}
	input := strings.TrimSpace(strings.ToLower(line))
	return input == "y" || input == "yes", nil
}
