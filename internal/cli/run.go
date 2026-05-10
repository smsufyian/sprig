package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	return &cobra.Command{
		Use:                "run <name> -- <command>",
		Short:              "Run a command inside a space",
		Args:               cobra.MinimumNArgs(1),
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
}
