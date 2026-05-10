package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newPruneCmd() *cobra.Command {
	var days int
	var dryRun bool

	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Remove stopped or stale spaces",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().IntVar(&days, "days", 3, "remove spaces stopped for more than N days")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "show what would be removed without acting")
	return cmd
}
