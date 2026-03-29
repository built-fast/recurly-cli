package cmd

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
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

	output := buf.String()
	if !strings.Contains(output, "# Recurly CLI Skill") {
		t.Error("expected SKILL.md header in output")
	}
	if !strings.Contains(output, "## Authentication") {
		t.Error("expected Authentication section in output")
	}
	if !strings.Contains(output, "## Command Reference") {
		t.Error("expected Command Reference section in output")
	}
}

func TestSkillInstallCmd(t *testing.T) {
	t.Parallel()

	tmpDir := t.TempDir()
	origFn := claudeSkillsDir
	claudeSkillsDir = func() (string, error) {
		return tmpDir, nil
	}
	t.Cleanup(func() { claudeSkillsDir = origFn })

	root := NewRootCmd()
	buf := new(bytes.Buffer)
	root.SetOut(buf)
	root.SetArgs([]string{"skill", "install"})

	if err := root.Execute(); err != nil {
		t.Fatalf("skill install command failed: %v", err)
	}

	dest := filepath.Join(tmpDir, "recurly", "SKILL.md")
	content, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading installed SKILL.md: %v", err)
	}

	if !strings.Contains(string(content), "# Recurly CLI Skill") {
		t.Error("installed SKILL.md missing expected header")
	}

	output := buf.String()
	if !strings.Contains(output, "Installed SKILL.md to") {
		t.Error("expected install confirmation message")
	}
}
