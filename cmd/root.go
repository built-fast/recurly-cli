package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// version is set via ldflags at build time:
//
//	go build -ldflags "-X github.com/built-fast/recurly-cli/cmd.version=v0.1.0"
var version = "dev"

// NewRootCmd creates and returns the root command with all subcommands registered.
func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "recurly",
		Short:   "Recurly CLI — manage Recurly resources from the command line",
		Version: version,
	}

	rootCmd.SetVersionTemplate("recurly-cli {{.Version}}\n")

	rootCmd.PersistentFlags().String("api-key", "", "Recurly API key")
	rootCmd.PersistentFlags().String("region", "us", "Recurly region (us or eu)")
	rootCmd.PersistentFlags().String("output", "table", "Output format (table, json, json-pretty)")

	_ = viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	_ = viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	rootCmd.AddCommand(newConfigureCmd())

	return rootCmd
}

// Execute runs the root command.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
