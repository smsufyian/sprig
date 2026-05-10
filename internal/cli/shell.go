package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newShellCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell <name>",
		Short: "Open an interactive shell inside a space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
}
