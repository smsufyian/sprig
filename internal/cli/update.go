package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newUpdateCmd() *cobra.Command {
	var targetVersion string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update sprig to the latest release",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().StringVar(&targetVersion, "version", "", "target version (default: latest)")
	return cmd
}
