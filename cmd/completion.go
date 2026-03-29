package cmd

import (
	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generate shell completion scripts",
		Long:  "Generate shell completion scripts for bash, zsh, fish, or powershell.",
	}

	completionCmd.AddCommand(newCompletionBashCmd())
	completionCmd.AddCommand(newCompletionZshCmd())
	completionCmd.AddCommand(newCompletionFishCmd())
	completionCmd.AddCommand(newCompletionPowershellCmd())

	return completionCmd
}

func newCompletionBashCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "bash",
		Short: "Generate bash completion script",
		Example: `  # Add to your ~/.bashrc:
  source <(recurly completion bash)

  # Or write to a file:
  recurly completion bash > /etc/bash_completion.d/recurly`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenBashCompletionV2(cmd.OutOrStdout(), true)
		},
	}
}

func newCompletionZshCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "zsh",
		Short: "Generate zsh completion script",
		Example: `  # Add to your ~/.zshrc (before compinit):
  source <(recurly completion zsh)

  # Or write to a file in your fpath:
  recurly completion zsh > "${fpath[1]}/_recurly"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenZshCompletion(cmd.OutOrStdout())
		},
	}
}

func newCompletionFishCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "fish",
		Short: "Generate fish completion script",
		Example: `  # Add to your fish config:
  recurly completion fish | source

  # Or write to a file:
  recurly completion fish > ~/.config/fish/completions/recurly.fish`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenFishCompletion(cmd.OutOrStdout(), true)
		},
	}
}

func newCompletionPowershellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "powershell",
		Short: "Generate PowerShell completion script",
		Example: `  # Add to your PowerShell profile:
  recurly completion powershell | Out-String | Invoke-Expression`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Root().GenPowerShellCompletionWithDesc(cmd.OutOrStdout())
		},
	}
}
