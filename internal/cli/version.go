package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/smsufyian/sprig/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Printf("sprig %s (commit %s, built %s)\n",
				version.Version, version.Commit, version.BuildDate)
			return nil
		},
	}
}
