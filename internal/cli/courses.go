package cli

import (
	"fmt"

	"cu-sync/internal/usecase/courses"

	"github.com/spf13/cobra"
)

var coursesCmd = &cobra.Command{
	Use:   "courses",
	Short: "List your courses",
	Long:  `Show all published courses with progress overview.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		client := mustClient()

		uc := courses.New(client)
		result, err := uc.List(ctx)
		if err != nil {
			fmt.Printf("Failed to fetch courses: %v\n", err)
			return
		}

		fmt.Printf("Your courses (%d)\n\n", len(result.Items))
		for i, course := range result.Items {
			fmt.Printf("  %d. [%d] %s\n", i+1, course.ID, course.Name)
		}
		fmt.Println("\nUse course ID or name with other commands:")
		fmt.Println("  cu deadlines go")
		fmt.Println("  cu grades алгоритмы")
		fmt.Println("  cu materials 901")
	},
}
