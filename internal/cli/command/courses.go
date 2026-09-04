package command

import (
	"fmt"

	"github.com/EgorTarasov/cu/internal/usecase/courses"

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
			exitErrf("Failed to fetch courses: %v", err)
		}

		fmt.Printf("Active (%d)\n\n", len(result.Active))
		for i, course := range result.Active {
			fmt.Printf("  %d. [%d] %s\n", i+1, course.ID, course.Name)
		}
		if len(result.Archived) > 0 {
			fmt.Printf("\nArchived (%d)\n\n", len(result.Archived))
			for i, course := range result.Archived {
				fmt.Printf("  %d. [%d] %s [archived]\n", i+1, course.ID, course.Name)
			}
		}
		fmt.Println("\nUse course ID or name with other commands:")
		fmt.Println("  cuni deadlines go")
		fmt.Println("  cuni grades алгоритмы")
		fmt.Println("  cuni materials 901")
	},
}
