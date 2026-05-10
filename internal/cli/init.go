package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Bootstrap sprig on this machine",
		Long:  "Creates the ~/.sprig directory and verifies platform requirements.",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir := sprigDir()
			dirs := []string{
				dir,
				filepath.Join(dir, "snapshots", "base"),
				filepath.Join(dir, "snapshots", "spaces"),
				filepath.Join(dir, "snapshots", "named"),
				filepath.Join(dir, "spaces"),
			}
			for _, d := range dirs {
				if err := os.MkdirAll(d, 0755); err != nil {
					return fmt.Errorf("create %s: %w", d, err)
				}
			}

			fmt.Println("sprig initialized.")
			fmt.Printf("  directory  %s\n", dir)
			fmt.Println()

			switch runtime.GOOS {
			case "darwin":
				fmt.Println("  macOS detected.")
				fmt.Println("  Lima VM + NixOS container support coming in a future release.")
			default:
				fmt.Println("  Linux detected.")
				fmt.Println("  NixOS container support coming in a future release.")
			}
			return nil
		},
	}
}
