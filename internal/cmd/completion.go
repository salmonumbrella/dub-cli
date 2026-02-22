// internal/cmd/completion.go
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for dub.

To load completions:

Bash:
  $ source <(dub completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ dub completion bash > /etc/bash_completion.d/dub
  # macOS:
  $ dub completion bash > $(brew --prefix)/etc/bash_completion.d/dub

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ dub completion zsh > "${fpath[1]}/_dub"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ dub completion fish | source

  # To load completions for each session, execute once:
  $ dub completion fish > ~/.config/fish/completions/dub.fish

PowerShell:
  PS> dub completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> dub completion powershell > dub.ps1
  # and source this file from your PowerShell profile.
`,
		DisableFlagsInUseLine: true,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return cmd.Root().GenBashCompletion(os.Stdout)
			case "zsh":
				return cmd.Root().GenZshCompletion(os.Stdout)
			case "fish":
				return cmd.Root().GenFishCompletion(os.Stdout, true)
			case "powershell":
				return cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
			}
			return nil
		},
	}

	return cmd
}
