package command

import (
	"fmt"

	"cu-sync/internal/version"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("cu %s (commit: %s, built: %s)\n",
			version.Version, version.Commit, version.Date)
	},
}

func init() {
	RootCmd.AddCommand(versionCmd)
}
