package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var coursesCmd = &cobra.Command{
	Use:   "courses",
	Short: "Sync courses (requires authentication)",
	Long:  `Synchronizes course data from Central University. Requires valid authentication tokens.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Course synchronization - Coming soon!")
		fmt.Println("This command will sync your course data from Central University.")
	},
}
