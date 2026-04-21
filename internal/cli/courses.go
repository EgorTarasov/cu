package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var coursesCmd = &cobra.Command{
	Use:   "courses",
	Short: "List your courses",
	Long:  `Show all published courses with progress overview.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		client := mustClient()

		courses, err := client.GetStudentCourses(ctx, maxCoursesLimit, "published")
		if err != nil {
			fmt.Printf("Failed to fetch courses: %v\n", err)
			return
		}

		fmt.Printf("Your courses (%d)\n\n", len(courses.Items))
		for i, course := range courses.Items {
			fmt.Printf("  %d. [%d] %s\n", i+1, course.ID, course.Name)
		}
		fmt.Println("\nUse course ID or name with other commands:")
		fmt.Println("  cu deadlines go")
		fmt.Println("  cu grades алгоритмы")
		fmt.Println("  cu materials 901")
	},
}
