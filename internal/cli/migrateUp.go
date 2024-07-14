package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Execute all of the pending migrations",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cli.app.Migrate().Up()
		if err != nil {
			log.Error().Err(err).Msg("failed to migrate up")
		}

		return nil
	},
}
