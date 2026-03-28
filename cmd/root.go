package cmd

import (
	"fmt"
	"os"

	"github.com/built-fast/recurly-cli/internal/client"
	"github.com/built-fast/recurly-cli/internal/config"
	"github.com/built-fast/recurly-cli/internal/output"
	"github.com/itchyny/gojq"
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
	rootCmd.SilenceErrors = true

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		output.SetJQ(nil)
		jqExpr, _ := cmd.Flags().GetString("jq")
		if jqExpr != "" {
			outputChanged := cmd.Flags().Changed("output")
			outputFormat, _ := cmd.Flags().GetString("output")

			if outputChanged && outputFormat == "table" {
				return fmt.Errorf("--jq and --output table are mutually exclusive")
			}

			if !outputChanged {
				viper.Set("output", "json")
			}

			query, err := gojq.Parse(jqExpr)
			if err != nil {
				return fmt.Errorf("invalid jq expression: %w", err)
			}

			code, err := gojq.Compile(query)
			if err != nil {
				return fmt.Errorf("compiling jq expression: %w", err)
			}

			output.SetJQ(code)
		}

		if err := config.Init(); err != nil {
			return err
		}
		if region := viper.GetString("region"); region != "" {
			if err := client.ValidateRegion(region); err != nil {
				return err
			}
		}

		return nil
	}

	rootCmd.PersistentFlags().String("api-key", "", "Recurly API key")
	rootCmd.PersistentFlags().String("jq", "", "Filter JSON output with a jq expression (built-in, no external jq required)")
	rootCmd.PersistentFlags().String("region", "us", "Recurly region (us or eu)")
	rootCmd.PersistentFlags().String("output", "table", "Output format (table, json, json-pretty)")

	_ = viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
	_ = viper.BindPFlag("region", rootCmd.PersistentFlags().Lookup("region"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))

	_ = viper.BindEnv("api_key", "RECURLY_API_KEY")
	_ = viper.BindEnv("region", "RECURLY_REGION")

	rootCmd.AddCommand(newConfigureCmd())
	rootCmd.AddCommand(newAccountsCmd())
	rootCmd.AddCommand(newPlansCmd())
	rootCmd.AddCommand(newItemsCmd())

	return rootCmd
}

// Execute runs the root command.
func Execute() {
	cmd := NewRootCmd()
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(cmd.ErrOrStderr(), client.FormatError(err))
		os.Exit(1)
	}
}
