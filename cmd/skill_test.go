package cmd

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/built-fast/recurly-cli/skills"
)

func TestSkillCmd(t *testing.T) {
	t.Parallel()

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"skill"})

	if err := root.Execute(); err != nil {
		t.Fatalf("skill command failed: %v", err)
	}

	// Verify byte-for-byte match with embedded content
	expected, err := skills.SkillMD()
	if err != nil {
		t.Fatalf("reading embedded SKILL.md: %v", err)
	}

	if !bytes.Equal(buf.Bytes(), expected) {
		t.Errorf("output does not match embedded SKILL.md byte-for-byte (got %d bytes, want %d bytes)", buf.Len(), len(expected))
	}
}

func testInstallConfig(t *testing.T) *installConfig {
	t.Helper()
	return &installConfig{
		agentsDir: filepath.Join(t.TempDir(), "agents"),
		claudeDir: filepath.Join(t.TempDir(), "claude"),
		version:   "v0.1.0-test",
		symlink:   os.Symlink,
	}
}

func TestRunSkillInstall_FileCreation(t *testing.T) {
	t.Parallel()

	cfg := testInstallConfig(t)
	buf := new(bytes.Buffer)

	if err := runSkillInstall(buf, cfg); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify SKILL.md written to agents dir
	dest := filepath.Join(cfg.agentsDir, "recurly", "SKILL.md")
	content, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading installed SKILL.md: %v", err)
	}

	expected, err := skills.SkillMD()
	if err != nil {
		t.Fatalf("reading embedded SKILL.md: %v", err)
	}
	if !bytes.Equal(content, expected) {
		t.Error("installed SKILL.md does not match embedded content")
	}

	// Verify .version file
	versionContent, err := os.ReadFile(filepath.Join(cfg.agentsDir, "recurly", ".version"))
	if err != nil {
		t.Fatalf("reading .version: %v", err)
	}
	if got := strings.TrimSpace(string(versionContent)); got != cfg.version {
		t.Errorf(".version = %q, want %q", got, cfg.version)
	}

	// Verify success output
	output := buf.String()
	if !strings.Contains(output, "Installed SKILL.md to") {
		t.Error("missing install confirmation")
	}
	if !strings.Contains(output, "Linked at") {
		t.Error("missing symlink confirmation")
	}
	if !strings.Contains(output, "recurly configure") {
		t.Error("missing next steps")
	}

	// Verify symlink in claude dir
	link := filepath.Join(cfg.claudeDir, "recurly", "SKILL.md")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("reading symlink: %v", err)
	}
	if target != dest {
		t.Errorf("symlink target = %q, want %q", target, dest)
	}
}

func TestRunSkillInstall_Symlink(t *testing.T) {
	t.Parallel()

	cfg := testInstallConfig(t)
	buf := new(bytes.Buffer)

	if err := runSkillInstall(buf, cfg); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	link := filepath.Join(cfg.claudeDir, "recurly", "SKILL.md")
	dest := filepath.Join(cfg.agentsDir, "recurly", "SKILL.md")

	// Verify it is a symlink pointing to the correct target
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat symlink: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("expected symlink, got regular file")
	}

	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("readlink: %v", err)
	}
	if target != dest {
		t.Errorf("symlink target = %q, want %q", target, dest)
	}

	// Verify the symlink resolves to correct content
	content, err := os.ReadFile(link)
	if err != nil {
		t.Fatalf("reading through symlink: %v", err)
	}
	expected, err := skills.SkillMD()
	if err != nil {
		t.Fatalf("reading embedded SKILL.md: %v", err)
	}
	if !bytes.Equal(content, expected) {
		t.Error("content through symlink does not match embedded SKILL.md")
	}
}

func TestRunSkillInstall_CopyFallback(t *testing.T) {
	t.Parallel()

	cfg := testInstallConfig(t)
	cfg.symlink = func(_, _ string) error {
		return errors.New("simulated symlink failure")
	}
	buf := new(bytes.Buffer)

	if err := runSkillInstall(buf, cfg); err != nil {
		t.Fatalf("install with copy fallback failed: %v", err)
	}

	// Verify file was copied (not symlinked)
	link := filepath.Join(cfg.claudeDir, "recurly", "SKILL.md")
	info, err := os.Lstat(link)
	if err != nil {
		t.Fatalf("lstat: %v", err)
	}
	if info.Mode()&os.ModeSymlink != 0 {
		t.Error("expected regular file (copy fallback), got symlink")
	}

	// Verify copied content matches
	content, err := os.ReadFile(link)
	if err != nil {
		t.Fatalf("reading copied file: %v", err)
	}
	expected, err := skills.SkillMD()
	if err != nil {
		t.Fatalf("reading embedded SKILL.md: %v", err)
	}
	if !bytes.Equal(content, expected) {
		t.Error("copied content does not match embedded SKILL.md")
	}

	// Verify agents dir still has the file
	agentFile := filepath.Join(cfg.agentsDir, "recurly", "SKILL.md")
	if _, err := os.Stat(agentFile); err != nil {
		t.Errorf("agents SKILL.md not found: %v", err)
	}

	// Verify output says "Copied" not "Linked"
	output := buf.String()
	if !strings.Contains(output, "Copied to") {
		t.Error("expected copy fallback message")
	}
	if strings.Contains(output, "Linked at") {
		t.Error("should not show symlink message when copy fallback was used")
	}
}

func TestRunSkillInstall_Idempotency(t *testing.T) {
	t.Parallel()

	cfg := testInstallConfig(t)

	// Run install twice
	for i := range 2 {
		buf := new(bytes.Buffer)
		if err := runSkillInstall(buf, cfg); err != nil {
			t.Fatalf("install run %d failed: %v", i+1, err)
		}
	}

	// Verify files are correct after second run
	dest := filepath.Join(cfg.agentsDir, "recurly", "SKILL.md")
	content, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading SKILL.md after second install: %v", err)
	}
	expected, err := skills.SkillMD()
	if err != nil {
		t.Fatalf("reading embedded SKILL.md: %v", err)
	}
	if !bytes.Equal(content, expected) {
		t.Error("SKILL.md content mismatch after second install")
	}

	// Verify symlink still works
	link := filepath.Join(cfg.claudeDir, "recurly", "SKILL.md")
	target, err := os.Readlink(link)
	if err != nil {
		t.Fatalf("reading symlink after second install: %v", err)
	}
	if target != dest {
		t.Errorf("symlink target after second install = %q, want %q", target, dest)
	}

	// Verify .version still present
	versionContent, err := os.ReadFile(filepath.Join(cfg.agentsDir, "recurly", ".version"))
	if err != nil {
		t.Fatalf("reading .version after second install: %v", err)
	}
	if got := strings.TrimSpace(string(versionContent)); got != cfg.version {
		t.Errorf(".version after second install = %q, want %q", got, cfg.version)
	}
}

func TestRunSkillUninstall_RemovesFiles(t *testing.T) {
	t.Parallel()

	cfg := testInstallConfig(t)
	buf := new(bytes.Buffer)

	// Install first
	if err := runSkillInstall(buf, cfg); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	// Verify files exist before uninstall
	agentsDir := filepath.Join(cfg.agentsDir, "recurly")
	claudeDir := filepath.Join(cfg.claudeDir, "recurly")
	if _, err := os.Stat(filepath.Join(agentsDir, "SKILL.md")); err != nil {
		t.Fatalf("SKILL.md not found before uninstall: %v", err)
	}
	if _, err := os.Stat(filepath.Join(claudeDir, "SKILL.md")); err != nil {
		t.Fatalf("claude SKILL.md not found before uninstall: %v", err)
	}

	// Uninstall
	buf.Reset()
	if err := runSkillUninstall(buf, cfg); err != nil {
		t.Fatalf("uninstall failed: %v", err)
	}

	// Verify directories removed
	if _, err := os.Stat(agentsDir); !os.IsNotExist(err) {
		t.Errorf("agents recurly dir still exists after uninstall")
	}
	if _, err := os.Stat(claudeDir); !os.IsNotExist(err) {
		t.Errorf("claude recurly dir still exists after uninstall")
	}

	// Verify success message
	if !strings.Contains(buf.String(), "Uninstalled") {
		t.Error("missing uninstall confirmation message")
	}
}

func TestRunSkillUninstall_NoOp(t *testing.T) {
	t.Parallel()

	cfg := testInstallConfig(t)
	buf := new(bytes.Buffer)

	// Uninstall without installing first — should be a no-op
	if err := runSkillUninstall(buf, cfg); err != nil {
		t.Fatalf("uninstall on non-existent dirs failed: %v", err)
	}

	// Verify success message still printed
	if !strings.Contains(buf.String(), "Uninstalled") {
		t.Error("missing uninstall confirmation message for no-op case")
	}
}

func TestRunSkillInstall_VersionStamp(t *testing.T) {
	t.Parallel()

	cfg := testInstallConfig(t)
	buf := new(bytes.Buffer)

	if err := runSkillInstall(buf, cfg); err != nil {
		t.Fatalf("install failed: %v", err)
	}

	versionFile := filepath.Join(cfg.agentsDir, "recurly", ".version")
	content, err := os.ReadFile(versionFile)
	if err != nil {
		t.Fatalf("reading .version: %v", err)
	}

	got := strings.TrimSpace(string(content))
	if got != cfg.version {
		t.Errorf(".version = %q, want %q", got, cfg.version)
	}

	// Verify version appears in output
	if !strings.Contains(buf.String(), cfg.version) {
		t.Error("version not shown in output")
	}
}
