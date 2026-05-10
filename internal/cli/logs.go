package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newLogsCmd() *cobra.Command {
	var follow bool

	cmd := &cobra.Command{
		Use:   "logs",
		Short: "Show sprig logs (~/.sprig/sprig.log)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "tail the log file")
	return cmd
}
