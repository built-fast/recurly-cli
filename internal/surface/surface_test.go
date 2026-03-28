package surface

import (
	"fmt"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestGenerate(t *testing.T) {
	root := &cobra.Command{Use: "myapp"}
	root.PersistentFlags().String("api-key", "", "API key")
	root.PersistentFlags().String("region", "us", "Region")

	sub := &cobra.Command{Use: "resources"}
	get := &cobra.Command{Use: "get <id>"}
	get.Flags().Bool("verbose", false, "Verbose output")

	sub.AddCommand(get)
	root.AddCommand(sub)

	out := Generate(root)

	// Must end with trailing newline
	if !strings.HasSuffix(out, "\n") {
		t.Error("output must end with trailing newline")
	}

	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")

	// Check sorted
	for i := 1; i < len(lines); i++ {
		if lines[i] < lines[i-1] {
			t.Errorf("lines not sorted: %q before %q", lines[i-1], lines[i])
		}
	}

	// Check expected entries exist
	expected := []string{
		"ARG myapp resources get <id>",
		"CMD myapp",
		"CMD myapp resources",
		"CMD myapp resources get",
		"FLAG myapp --api-key string",
		"FLAG myapp --region string",
		"FLAG myapp resources get --verbose bool",
	}
	for _, e := range expected {
		found := false
		for _, l := range lines {
			if l == e {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected line not found: %q", e)
		}
	}

	// Persistent flags must NOT appear on children
	for _, l := range lines {
		if strings.Contains(l, "FLAG myapp resources --api-key") {
			t.Error("persistent flag --api-key duplicated on child command")
		}
		if strings.Contains(l, "FLAG myapp resources get --api-key") {
			t.Error("persistent flag --api-key duplicated on grandchild command")
		}
	}

	// --help must not appear
	for _, l := range lines {
		if strings.Contains(l, "--help") {
			t.Error("--help flag should be excluded")
		}
	}

	// help command must not appear
	for _, l := range lines {
		if strings.Contains(l, "CMD myapp help") {
			t.Error("help command should be excluded")
		}
	}

	fmt.Println(out)
}

func TestGenerateExcludesCompletion(t *testing.T) {
	root := &cobra.Command{Use: "myapp"}
	root.AddCommand(&cobra.Command{Use: "completion"})
	root.AddCommand(&cobra.Command{Use: "real"})

	out := Generate(root)
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")

	for _, l := range lines {
		if strings.Contains(l, "completion") {
			t.Error("completion command should be excluded")
		}
	}

	found := false
	for _, l := range lines {
		if l == "CMD myapp real" {
			found = true
		}
	}
	if !found {
		t.Error("expected CMD myapp real")
	}
}
