package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newConfigureCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "configure",
		Short: "Configure Recurly CLI settings",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "configure command not yet implemented")
		},
	}
}
