package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newOpenCmd() *cobra.Command {
	var service string

	cmd := &cobra.Command{
		Use:   "open <name>",
		Short: "Print the URLs and ports for a space's services",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().StringVar(&service, "service", "", "print only this service's URL")
	return cmd
}
