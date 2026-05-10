package db

import (
	"fmt"

	"github.com/spf13/cobra"
)

// NewDBCmd returns the `sprig db` subcommand group.
func NewDBCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db",
		Short: "Manage database snapshots and seeding",
		Long:  "Pull from production, seed from a file, generate synthetic data, and manage snapshots.",
	}
	cmd.AddCommand(
		newPullCmd(),
		newSeedCmd(),
		newGenerateCmd(),
		newResetCmd(),
		newSnapshotCmd(),
		newRestoreCmd(),
	)
	return cmd
}

func newPullCmd() *cobra.Command {
	var from string
	cmd := &cobra.Command{
		Use:   "pull",
		Short: "Snapshot a production database into the base snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "connection string or named connection (required)")
	_ = cmd.MarkFlagRequired("from")
	return cmd
}

func newSeedCmd() *cobra.Command {
	var file string
	var setAsBase bool
	cmd := &cobra.Command{
		Use:   "seed <space>",
		Short: "Seed a space's database from a dump file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().StringVar(&file, "file", "", "path to .sql or .dump file (required)")
	cmd.Flags().BoolVar(&setAsBase, "set-as-base", false, "promote this seed to the base snapshot")
	_ = cmd.MarkFlagRequired("file")
	return cmd
}

func newGenerateCmd() *cobra.Command {
	var rows int
	cmd := &cobra.Command{
		Use:   "generate <space>",
		Short: "Generate synthetic data for a space's database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().IntVar(&rows, "rows", 1000, "approximate number of rows per table")
	return cmd
}

func newResetCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reset <space>",
		Short: "Re-clone the space's database from the base snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
}

func newSnapshotCmd() *cobra.Command {
	var label string
	cmd := &cobra.Command{
		Use:   "snapshot <space>",
		Short: "Save a named snapshot of the space's database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().StringVar(&label, "as", "", "label for the snapshot (required)")
	_ = cmd.MarkFlagRequired("as")
	return cmd
}

func newRestoreCmd() *cobra.Command {
	var from string
	cmd := &cobra.Command{
		Use:   "restore <space>",
		Short: "Restore the space's database from a named snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Println("coming in a future release")
			return nil
		},
	}
	cmd.Flags().StringVar(&from, "from", "", "snapshot label to restore from (required)")
	_ = cmd.MarkFlagRequired("from")
	return cmd
}
