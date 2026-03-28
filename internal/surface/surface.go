package surface

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// Generate walks the Cobra command tree and produces a deterministic text
// snapshot of the CLI surface. Each line is one of three space-delimited
// entry types: ARG, CMD, or FLAG. Lines are sorted alphabetically and the
// output ends with a trailing newline.
func Generate(cmd *cobra.Command) string {
	var lines []string
	walk(cmd, &lines)
	sort.Strings(lines)
	return strings.Join(lines, "\n") + "\n"
}

func walk(cmd *cobra.Command, lines *[]string) {
	path := cmd.CommandPath()

	*lines = append(*lines, "CMD "+path)

	for _, arg := range parseArgs(cmd.Use) {
		*lines = append(*lines, fmt.Sprintf("ARG %s %s", path, arg))
	}

	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}
		*lines = append(*lines, fmt.Sprintf("FLAG %s --%s %s", path, f.Name, f.Value.Type()))
	})

	for _, child := range cmd.Commands() {
		if child.Name() == "help" || child.Name() == "completion" {
			continue
		}
		walk(child, lines)
	}
}

func parseArgs(use string) []string {
	parts := strings.Fields(use)
	if len(parts) <= 1 {
		return nil
	}
	return parts[1:]
}
