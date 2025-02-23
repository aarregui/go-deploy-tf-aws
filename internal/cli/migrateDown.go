package cli

import (
	"log/slog"

	"github.com/spf13/cobra"
)

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback to the previous migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cli.app.Migrate().Steps(-1)
		if err != nil {
			slog.Error("failed to migrate down", "error", err)
		}

		return nil
	},
}
