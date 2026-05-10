package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDestroyCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "destroy <name>",
		Short: "Destroy a space and release all its resources",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
}
