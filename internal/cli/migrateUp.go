package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Execute all of the pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cli.app.Migrate().Up()
		if err != nil {
			slog.Error("failed to migrate up", "error", err)
		}

		return nil
	},
}
