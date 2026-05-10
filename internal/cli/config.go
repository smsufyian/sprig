package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
)

func newConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage sprig configuration",
	}
	cmd.AddCommand(newConfigEditCmd())
	return cmd
}

func newConfigEditCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "edit",
		Short: "Open sprig.toml in $EDITOR (creates it with a template if missing)",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			tomlPath := filepath.Join(cwd, "sprig.toml")

			if _, err := os.Stat(tomlPath); os.IsNotExist(err) {
				template := `# sprig.toml — override auto-detected settings
# Only include fields you want to change.
#
# [project]
# name = "my-app"
# run  = "python manage.py runserver"
#
# [services.postgres]
# version = "16"
#
# [services.redis]
# enabled = true
#
# [services.kafka]
# version = "3.7"
# partitions = 3
#
# [services.app]
# port = 8080
# env  = { LOG_LEVEL = "debug", FEATURE_X = "true" }
#
# [db]
# migrations = "./db/migrations"
`
				if err := os.WriteFile(tomlPath, []byte(template), 0644); err != nil {
					return fmt.Errorf("create sprig.toml: %w", err)
				}
				fmt.Println("Created sprig.toml with template.")
			}

			editor := os.Getenv("EDITOR")
			if editor == "" {
				editor = "vi"
			}
			c := exec.Command(editor, tomlPath)
			c.Stdin = os.Stdin
			c.Stdout = os.Stdout
			c.Stderr = os.Stderr
			return c.Run()
		},
	}
}
