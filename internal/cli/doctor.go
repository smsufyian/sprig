package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose platform issues",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
}
