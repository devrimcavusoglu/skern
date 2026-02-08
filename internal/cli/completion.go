package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion [bash|zsh|fish]",
		Short:     "Generate shell completion scripts",
		Long:      "Generate shell completion scripts for bash, zsh, or fish.\n\nLoad completions in your shell session:\n\n  bash:  source <(scribe completion bash)\n  zsh:   source <(scribe completion zsh)\n  fish:  scribe completion fish | source",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			shell := args[0]
			root := cmd.Root()

			switch shell {
			case "bash":
				return root.GenBashCompletion(cmd.OutOrStdout())
			case "zsh":
				return root.GenZshCompletion(cmd.OutOrStdout())
			case "fish":
				return root.GenFishCompletion(cmd.OutOrStdout(), true)
			default:
				return &ValidationError{Message: fmt.Sprintf("unsupported shell %q: must be bash, zsh, or fish", shell)}
			}
		},
	}

	return cmd
}
