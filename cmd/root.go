package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var version = "dev" // overridden by ldflags

var rootCmd = &cobra.Command{
	Use:     "recurly",
	Short:   "Recurly CLI — manage Recurly resources from the command line",
	Version: version,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().String("api-key", "", "Recurly API key")
	rootCmd.PersistentFlags().String("region", "us", "Recurly region (us or eu)")
	rootCmd.PersistentFlags().String("output", "table", "Output format (table, json, json-pretty)")

	_ = viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	_ = viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
}
