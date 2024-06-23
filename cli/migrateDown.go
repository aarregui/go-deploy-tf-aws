package cli

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var migrateDownCmd = &cobra.Command{
	Use:   "down",
	Short: "Rollback to the previous migration",
	RunE: func(cmd *cobra.Command, args []string) error {
		err := cli.app.Migrate().Steps(-1)
		if err != nil {
			log.Error().Err(err).Msg("failed to migrate down")
		}

		return nil
	},
}
