package cmd

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

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
	cmd.AddCommand(newSkillUninstallCmd())

	return cmd
}

// installConfig holds dependencies for skill installation, enabling testability.
type installConfig struct {
	agentsDir string
	claudeDir string
	version   string
	symlink   func(oldname, newname string) error
}

func defaultInstallConfig() (*installConfig, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("determining home directory: %w", err)
	}
	return &installConfig{
		agentsDir: filepath.Join(home, ".agents", "skills"),
		claudeDir: filepath.Join(home, ".claude", "skills"),
		version:   version,
		symlink:   os.Symlink,
	}, nil
}

func newSkillInstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "install",
		Short: "Install SKILL.md to the agents skills directory with Claude Code symlink",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := defaultInstallConfig()
			if err != nil {
				return err
			}
			return runSkillInstall(cmd.OutOrStdout(), cfg)
		},
	}
}

func newSkillUninstallCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "uninstall",
		Short: "Remove installed skill files",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := defaultInstallConfig()
			if err != nil {
				return err
			}
			return runSkillUninstall(cmd.OutOrStdout(), cfg)
		},
	}
}

// RefreshSkillsIfVersionChanged silently re-installs SKILL.md when the CLI
// version has changed since the last install. It is best-effort: no output on
// success, no error on failure. It is a no-op if the skill was never installed
// or the build is a dev build.
func RefreshSkillsIfVersionChanged() {
	cfg, err := defaultInstallConfig()
	if err != nil {
		return
	}
	refreshSkills(cfg)
}

func refreshSkills(cfg *installConfig) {
	// Skip dev builds
	if cfg.version == "" || cfg.version == "dev" {
		return
	}

	versionFile := filepath.Join(cfg.agentsDir, "recurly", ".version")
	data, err := os.ReadFile(versionFile)
	if err != nil {
		// .version missing → never installed → no-op
		return
	}

	installed := strings.TrimSpace(string(data))
	if installed == cfg.version {
		// Already up to date
		return
	}

	// Version mismatch — silently re-install
	_ = runSkillInstall(io.Discard, cfg)
}

func runSkillUninstall(w io.Writer, cfg *installConfig) error {
	// Remove ~/.agents/skills/recurly/
	agentsRecurly := filepath.Join(cfg.agentsDir, "recurly")
	if err := os.RemoveAll(agentsRecurly); err != nil {
		return fmt.Errorf("removing %s: %w", agentsRecurly, err)
	}

	// Remove ~/.claude/skills/recurly/
	claudeRecurly := filepath.Join(cfg.claudeDir, "recurly")
	if err := os.RemoveAll(claudeRecurly); err != nil {
		return fmt.Errorf("removing %s: %w", claudeRecurly, err)
	}

	_, _ = fmt.Fprintln(w, "Uninstalled recurly skill files.")
	return nil
}

func runSkillInstall(w io.Writer, cfg *installConfig) error {
	content, err := skills.SkillMD()
	if err != nil {
		return fmt.Errorf("reading embedded SKILL.md: %w", err)
	}

	// 1. Write to ~/.agents/skills/recurly/SKILL.md
	destDir := filepath.Join(cfg.agentsDir, "recurly")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("creating agents skills directory: %w", err)
	}

	dest := filepath.Join(destDir, "SKILL.md")
	if err := os.WriteFile(dest, content, 0o644); err != nil {
		return fmt.Errorf("writing SKILL.md: %w", err)
	}

	// 2. Write .version file
	versionFile := filepath.Join(destDir, ".version")
	if err := os.WriteFile(versionFile, []byte(cfg.version+"\n"), 0o644); err != nil {
		return fmt.Errorf("writing .version: %w", err)
	}

	// 3. Symlink from ~/.claude/skills/recurly/SKILL.md → installed file
	linkDir := filepath.Join(cfg.claudeDir, "recurly")
	if err := os.MkdirAll(linkDir, 0o755); err != nil {
		return fmt.Errorf("creating Claude skills directory: %w", err)
	}

	link := filepath.Join(linkDir, "SKILL.md")
	// Remove existing file/symlink for idempotency
	_ = os.Remove(link)

	usedCopy := false
	if err := cfg.symlink(dest, link); err != nil {
		// Fall back to file copy if symlink fails (e.g., cross-device)
		if err := os.WriteFile(link, content, 0o644); err != nil {
			return fmt.Errorf("copying SKILL.md to Claude skills directory: %w", err)
		}
		usedCopy = true
	}

	// 4. Print success
	_, _ = fmt.Fprintf(w, "Installed SKILL.md to %s\n", dest)
	_, _ = fmt.Fprintf(w, "Version:  %s\n", cfg.version)
	if usedCopy {
		_, _ = fmt.Fprintf(w, "Copied to %s (symlink not supported)\n", link)
	} else {
		_, _ = fmt.Fprintf(w, "Linked at %s\n", link)
	}

	// 5. Next steps
	_, _ = fmt.Fprintln(w, "")
	_, _ = fmt.Fprintln(w, "Next steps:")
	_, _ = fmt.Fprintln(w, "  Run `recurly configure` to set up authentication")

	return nil
}
