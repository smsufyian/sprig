package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCreateCmd() *cobra.Command {
	var from string
	var ciMode bool
	var timeout string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a new isolated space",
		Long:  "Auto-detects your stack and creates a fully isolated environment with its own database and services.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}

	cmd.Flags().StringVar(&from, "from", "", "seed source: a named connection or 'production'")
	cmd.Flags().BoolVar(&ciMode, "ci", false, "output env vars for eval $(...) usage in CI")
	cmd.Flags().StringVar(&timeout, "timeout", "60s", "max time to wait for the space to be ready")
	return cmd
}
