package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// newTestCmd returns a cobra.Command wired to the given stdin, stdout, and stderr buffers.
func newTestCmd(stdin *bytes.Buffer, stdout, stderr *bytes.Buffer) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	cmd.SetIn(stdin)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	return cmd
}

func TestConfirm_AcceptsYes(t *testing.T) {
	t.Parallel()
	accepted := []string{"y", "yes", "Y", "YES", "Yes", "yEs"}
	for _, input := range accepted {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			stdin := bytes.NewBufferString(input + "\n")
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			cmd := newTestCmd(stdin, stdout, stderr)

			ok, err := confirm(cmd, "Continue? [y/N] ")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !ok {
				t.Errorf("expected confirm to return true for input %q", input)
			}
			if !strings.Contains(stderr.String(), "Continue? [y/N] ") {
				t.Errorf("expected prompt in stderr, got %q", stderr.String())
			}
		})
	}
}

func TestConfirm_RejectsNo(t *testing.T) {
	t.Parallel()
	rejected := []string{"n", "no", "N", "NO", "No"}
	for _, input := range rejected {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			stdin := bytes.NewBufferString(input + "\n")
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			cmd := newTestCmd(stdin, stdout, stderr)

			ok, err := confirm(cmd, "Continue? [y/N] ")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Errorf("expected confirm to return false for input %q", input)
			}
		})
	}
}

func TestConfirm_RejectsEmptyInput(t *testing.T) {
	t.Parallel()
	stdin := bytes.NewBufferString("\n")
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := newTestCmd(stdin, stdout, stderr)

	ok, err := confirm(cmd, "Continue? [y/N] ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Error("expected confirm to return false for empty input")
	}
}

func TestConfirm_RejectsRandomText(t *testing.T) {
	t.Parallel()
	inputs := []string{"maybe", "sure", "ok", "1", "true", "yep"}
	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			t.Parallel()
			stdin := bytes.NewBufferString(input + "\n")
			stdout := new(bytes.Buffer)
			stderr := new(bytes.Buffer)
			cmd := newTestCmd(stdin, stdout, stderr)

			ok, err := confirm(cmd, "Continue? [y/N] ")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ok {
				t.Errorf("expected confirm to return false for input %q", input)
			}
		})
	}
}

func TestConfirm_EOF_ReturnsError(t *testing.T) {
	t.Parallel()
	// Empty buffer with no newline simulates EOF
	stdin := bytes.NewBufferString("")
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	cmd := newTestCmd(stdin, stdout, stderr)

	_, err := confirm(cmd, "Continue? [y/N] ")
	if err == nil {
		t.Fatal("expected error on EOF, got nil")
	}
	if !strings.Contains(err.Error(), "reading confirmation") {
		t.Errorf("expected error to mention 'reading confirmation', got %q", err.Error())
	}
}
