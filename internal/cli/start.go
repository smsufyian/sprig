package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newStartCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "start <name>",
		Short: "Start a stopped space",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
}
