package cli

import (
	"os"

	"github.com/spf13/cobra"
)

func newCompletionCmd() *cobra.Command {
	return &cobra.Command{
		Use:       "completion bash|zsh|fish",
		Short:     "Print shell completion script",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash", "zsh", "fish"},
		RunE: func(cmd *cobra.Command, args []string) error {
			root := cmd.Root()
			switch args[0] {
			case "bash":
				return root.GenBashCompletion(os.Stdout)
			case "zsh":
				return root.GenZshCompletion(os.Stdout)
			case "fish":
				return root.GenFishCompletion(os.Stdout, true)
			}
			return nil
		},
	}
}
