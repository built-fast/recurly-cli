package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/built-fast/recurly-cli/skills"
	"github.com/spf13/cobra"
)

func newSkillCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "skill",
		Short: "Print the SKILL.md agent reference document",
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := skills.SkillMD()
			if err != nil {
				return fmt.Errorf("reading embedded SKILL.md: %w", err)
			}
			_, err = cmd.OutOrStdout().Write(content)
			return err
		},
	}

	cmd.AddCommand(newSkillInstallCmd())

	return cmd
}

// claudeSkillsDir returns the Claude Code skills directory path.
// Declared as a variable for testability.
var claudeSkillsDir = func() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("determining home directory: %w", err)
	}
	return filepath.Join(home, ".claude", "skills"), nil
}

func newSkillInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install SKILL.md to the Claude Code skills directory",
		RunE: func(cmd *cobra.Command, args []string) error {
			content, err := skills.SkillMD()
			if err != nil {
				return fmt.Errorf("reading embedded SKILL.md: %w", err)
			}

			dir, err := claudeSkillsDir()
			if err != nil {
				return err
			}

			destDir := filepath.Join(dir, "recurly")
			if err := os.MkdirAll(destDir, 0o755); err != nil {
				return fmt.Errorf("creating skills directory: %w", err)
			}

			dest := filepath.Join(destDir, "SKILL.md")
			if err := os.WriteFile(dest, content, 0o644); err != nil {
				return fmt.Errorf("writing SKILL.md: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Installed SKILL.md to %s\n", dest)
			return nil
		},
	}
}
