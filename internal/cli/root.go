package cli

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/smsufyian/sprig/internal/cli/db"
)

var outputMode string
var debugMode bool

// NewRootCmd builds and returns the root sprig command with all subcommands attached.
func NewRootCmd(ctx context.Context) *cobra.Command {
	root := &cobra.Command{
		Use:   "sprig",
		Short: "Isolated virtual spaces for development, testing, and CI",
		Long: `sprig creates fully isolated environments — each with its own application
stack, services (Postgres, Redis, Kafka, ...), and a copy of your production
data. Works completely offline. Complexity stays hidden.

  sprig create my-feature          spin up an isolated space
  sprig list                       see all your spaces
  sprig shell my-feature           jump into a space
  sprig destroy my-feature         tear it down`,
	}

	root.PersistentFlags().StringVar(&outputMode, "output", "text", "output format: text or json")
	root.PersistentFlags().BoolVar(&debugMode, "debug", false, "enable debug logging to ~/.sprig/sprig.log")
	root.SetContext(ctx)

	root.AddCommand(
		newInitCmd(),
		newCreateCmd(),
		newListCmd(),
		newStatusCmd(),
		newStartCmd(),
		newStopCmd(),
		newDestroyCmd(),
		newOpenCmd(),
		newShellCmd(),
		newRunCmd(),
		newConfigCmd(),
		newVersionCmd(),
		newCompletionCmd(),
		newPruneCmd(),
		newUpdateCmd(),
		newTelemetryCmd(),
		newDoctorCmd(),
		newLogsCmd(),
		db.NewDBCmd(),
	)

	return root
}
