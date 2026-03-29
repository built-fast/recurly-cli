package cmd

import (
	"bytes"
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
