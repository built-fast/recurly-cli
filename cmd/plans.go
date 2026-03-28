package cmd

import (
	"github.com/spf13/cobra"
)

func newPlansCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plans",
		Short: "Manage plans",
	}
	return cmd
}
