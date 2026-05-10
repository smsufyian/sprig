package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newTelemetryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Manage usage telemetry",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "status",
			Short: "Show telemetry opt-in/out status",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("Telemetry: enabled (default)")
				fmt.Println("  To disable: sprig telemetry disable")
				return nil
			},
		},
		&cobra.Command{
			Use:   "disable",
			Short: "Opt out of usage telemetry",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("coming in a future release")
				return nil
			},
		},
		&cobra.Command{
			Use:   "enable",
			Short: "Opt in to usage telemetry",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Println("coming in a future release")
				return nil
			},
		},
	)
	return cmd
}
