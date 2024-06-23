package cli

import (
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Manage the database state",
}

func init() {
	migrateCmd.AddCommand(migrateDownCmd)
	migrateCmd.AddCommand(migrateUpCmd)
}
