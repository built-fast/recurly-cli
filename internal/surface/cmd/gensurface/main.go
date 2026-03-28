package main

import (
	"fmt"

	"github.com/built-fast/recurly-cli/cmd"
	"github.com/built-fast/recurly-cli/internal/surface"
)

func main() {
	root := cmd.NewRootCmd()
	fmt.Print(surface.Generate(root))
}
