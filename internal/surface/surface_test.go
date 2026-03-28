package surface

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/built-fast/recurly-cli/cmd"
	"github.com/spf13/cobra"
)

func TestGoldenSurface(t *testing.T) {
	surfacePath := filepath.Join("..", "..", ".surface")

	expected, err := os.ReadFile(surfacePath)
	if err != nil {
		if os.IsNotExist(err) {
			t.Fatal(".surface file not found — run: make surface")
		}
		t.Fatalf("failed to read .surface: %v", err)
	}

	root := cmd.NewRootCmd()
	actual := Generate(root)

	if string(expected) != actual {
		t.Fatalf("CLI surface has changed. If intentional, run: make surface\n\nExpected:\n%s\nActual:\n%s", string(expected), actual)
	}
}

// helper to check a line exists in the output
func assertContains(t *testing.T, output, line string) {
	t.Helper()
	for _, l := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		if l == line {
			return
		}
	}
	t.Errorf("expected line not found: %q\noutput:\n%s", line, output)
}

// helper to check a line does NOT exist in the output
func assertNotContains(t *testing.T, output, substr string) {
	t.Helper()
	for _, l := range strings.Split(strings.TrimSuffix(output, "\n"), "\n") {
		if strings.Contains(l, substr) {
			t.Errorf("unexpected line containing %q found: %q", substr, l)
			return
		}
	}
}

func TestGenerate_BasicCommandTree(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	accounts := &cobra.Command{Use: "accounts"}
	list := &cobra.Command{Use: "list"}
	get := &cobra.Command{Use: "get <id>"}

	accounts.AddCommand(list, get)
	root.AddCommand(accounts)

	out := Generate(root)

	assertContains(t, out, "CMD cli")
	assertContains(t, out, "CMD cli accounts")
	assertContains(t, out, "CMD cli accounts list")
	assertContains(t, out, "CMD cli accounts get")
}

func TestGenerate_PositionalArgsSingleArg(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	get := &cobra.Command{Use: "get <account_id>"}
	root.AddCommand(get)

	out := Generate(root)

	assertContains(t, out, "ARG cli get <account_id>")
}

func TestGenerate_PositionalArgsMultipleArgs(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	move := &cobra.Command{Use: "move <source> <destination>"}
	root.AddCommand(move)

	out := Generate(root)

	assertContains(t, out, "ARG cli move <source>")
	assertContains(t, out, "ARG cli move <destination>")
}

func TestGenerate_NoArgsProducesNoARGEntries(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	list := &cobra.Command{Use: "list"}
	root.AddCommand(list)

	out := Generate(root)

	assertNotContains(t, out, "ARG")
}

func TestGenerate_LocalFlagsWithTypes(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	list := &cobra.Command{Use: "list"}
	list.Flags().String("filter", "", "filter expression")
	list.Flags().Bool("verbose", false, "verbose output")
	list.Flags().Int("limit", 20, "page limit")
	root.AddCommand(list)

	out := Generate(root)

	assertContains(t, out, "FLAG cli list --filter string")
	assertContains(t, out, "FLAG cli list --verbose bool")
	assertContains(t, out, "FLAG cli list --limit int")
}

func TestGenerate_PersistentFlagsOnlyOnDefiningCommand(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	root.PersistentFlags().String("api-key", "", "API key")
	root.PersistentFlags().String("region", "us", "Region")

	child := &cobra.Command{Use: "sub"}
	grandchild := &cobra.Command{Use: "deep"}
	child.AddCommand(grandchild)
	root.AddCommand(child)

	out := Generate(root)

	// Persistent flags appear on root
	assertContains(t, out, "FLAG cli --api-key string")
	assertContains(t, out, "FLAG cli --region string")

	// Persistent flags must NOT appear on children
	assertNotContains(t, out, "FLAG cli sub --api-key")
	assertNotContains(t, out, "FLAG cli sub --region")
	assertNotContains(t, out, "FLAG cli sub deep --api-key")
	assertNotContains(t, out, "FLAG cli sub deep --region")
}

func TestGenerate_ExcludesHelpCommand(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	root.AddCommand(&cobra.Command{Use: "real"})
	// Cobra auto-adds a help command; also manually add one to be explicit
	root.AddCommand(&cobra.Command{Use: "help"})

	out := Generate(root)

	assertNotContains(t, out, "CMD cli help")
	assertContains(t, out, "CMD cli real")
}

func TestGenerate_ExcludesCompletionCommand(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	root.AddCommand(&cobra.Command{Use: "completion"})
	root.AddCommand(&cobra.Command{Use: "real"})

	out := Generate(root)

	assertNotContains(t, out, "completion")
	assertContains(t, out, "CMD cli real")
}

func TestGenerate_ExcludesHelpFlag(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	sub := &cobra.Command{Use: "sub"}
	sub.Flags().String("output", "", "output format")
	root.AddCommand(sub)

	out := Generate(root)

	assertNotContains(t, out, "--help")
	// Non-help flags still appear
	assertContains(t, out, "FLAG cli sub --output string")
}

func TestGenerate_OutputSortedAlphabetically(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	root.PersistentFlags().String("api-key", "", "")

	beta := &cobra.Command{Use: "beta"}
	beta.Flags().Bool("zflag", false, "")
	beta.Flags().String("aflag", "", "")

	alpha := &cobra.Command{Use: "alpha <id>"}

	root.AddCommand(beta, alpha)

	out := Generate(root)
	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")

	for i := 1; i < len(lines); i++ {
		if lines[i] < lines[i-1] {
			t.Errorf("lines not sorted: %q before %q", lines[i-1], lines[i])
		}
	}
}

func TestGenerate_TrailingNewline(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	out := Generate(root)

	if !strings.HasSuffix(out, "\n") {
		t.Error("output must end with trailing newline")
	}
}

func TestGenerate_RootOnlyCommand(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	out := Generate(root)

	lines := strings.Split(strings.TrimSuffix(out, "\n"), "\n")
	if len(lines) != 1 {
		t.Errorf("expected 1 line for root-only command, got %d:\n%s", len(lines), out)
	}
	assertContains(t, out, "CMD cli")
}

func TestGenerate_DeepNesting(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	l1 := &cobra.Command{Use: "level1"}
	l2 := &cobra.Command{Use: "level2"}
	l3 := &cobra.Command{Use: "level3 <arg>"}
	l3.Flags().Int("count", 0, "count")

	l2.AddCommand(l3)
	l1.AddCommand(l2)
	root.AddCommand(l1)

	out := Generate(root)

	assertContains(t, out, "CMD cli")
	assertContains(t, out, "CMD cli level1")
	assertContains(t, out, "CMD cli level1 level2")
	assertContains(t, out, "CMD cli level1 level2 level3")
	assertContains(t, out, "ARG cli level1 level2 level3 <arg>")
	assertContains(t, out, "FLAG cli level1 level2 level3 --count int")
}

func TestGenerate_ChildWithOwnPersistentFlags(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	sub := &cobra.Command{Use: "sub"}
	sub.PersistentFlags().String("format", "json", "output format")

	leaf := &cobra.Command{Use: "leaf"}
	sub.AddCommand(leaf)
	root.AddCommand(sub)

	out := Generate(root)

	// Persistent flag defined on sub should appear on sub
	assertContains(t, out, "FLAG cli sub --format string")
	// But not on leaf
	assertNotContains(t, out, "FLAG cli sub leaf --format")
}

func TestGenerate_MixedFlagTypes(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	cmd := &cobra.Command{Use: "do"}
	cmd.Flags().Float64("rate", 0.0, "rate")
	cmd.Flags().StringSlice("tags", nil, "tags")
	cmd.Flags().Int64("big-num", 0, "big number")
	root.AddCommand(cmd)

	out := Generate(root)

	assertContains(t, out, "FLAG cli do --rate float64")
	assertContains(t, out, "FLAG cli do --tags stringSlice")
	assertContains(t, out, "FLAG cli do --big-num int64")
}

func TestGenerate_ExcludesHelpAndCompletionAtNestedLevels(t *testing.T) {
	root := &cobra.Command{Use: "cli"}
	sub := &cobra.Command{Use: "sub"}
	sub.AddCommand(&cobra.Command{Use: "help"})
	sub.AddCommand(&cobra.Command{Use: "completion"})
	sub.AddCommand(&cobra.Command{Use: "action"})
	root.AddCommand(sub)

	out := Generate(root)

	assertNotContains(t, out, "CMD cli sub help")
	assertNotContains(t, out, "CMD cli sub completion")
	assertContains(t, out, "CMD cli sub action")
}
