package command

import (
	"cu-sync/internal/update"

	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "cu",
	Short: "Central University CLI Tool",
	Long: `CU is a command-line tool for interacting with Central University services.
It provides access to courses, authentication, and data synchronization.`,
	PersistentPostRun: func(_ *cobra.Command, _ []string) {
		update.CheckForUpdate()
	},
}
