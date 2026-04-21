package cli

import (
	"fmt"
	"path/filepath"
	"strconv"

	"cu-sync/internal/cu"

	"github.com/spf13/cobra"
)

func init() {
	fetchCourseCmd.Flags().String("path", ".", "path to save the course data")
	fetchCourseCmd.Flags().Bool("dump", false, "dumps all course data")

	fetchCmd.AddCommand(fetchCourseCmd)
	fetchCmd.AddCommand(fetchCoursesCmd)
}

var fetchCmd = &cobra.Command{
	Use:   "fetch",
	Short: "Fetch data from Central University",
	Long:  `Fetch various data from Central University using authenticated requests.`,
}

var fetchCourseCmd = &cobra.Command{
	Use:   "course [course-id]",
	Short: "Fetch course overview by ID",
	Long:  `Fetch detailed course overview from Central University by course ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		courseID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Printf("Invalid course ID '%s': %v\n", args[0], err)
			return
		}

		fmt.Printf("Fetching course %d...\n", courseID)
		course, err := client.GetCourseOverview(ctx, courseID)
		if err != nil {
			fmt.Printf("Failed to fetch course: %v\n", err)
			return
		}

		fmt.Printf("Course: %s (ID: %d)\n", course.Name, course.ID)
		fmt.Printf("State: %s | Archived: %v\n", course.State, course.IsArchived)
		fmt.Printf("Themes: %d\n\n", len(course.Themes))

		dump, _ := cmd.Flags().GetBool("dump")

		if dump {
			basePath, _ := cmd.Flags().GetString("path")
			courseDir := filepath.Join(basePath, sanitizeFilename(course.Name)+strconv.Itoa(courseID))
			err = dumpCourse(ctx, client, course, courseDir)
			if err != nil {
				fmt.Printf("Failed to download course data: %v\n", err)
			}
		} else {
			for _, theme := range course.Themes {
				fmt.Printf("  %d. %s\n", theme.Order, theme.Name)
				for _, longread := range theme.Longreads {
					fmt.Printf("     - %s (%s)\n", longread.Name, longread.Type)
					if len(longread.Exercises) > 0 {
						fmt.Printf("       exercises: %d\n", len(longread.Exercises))
					}
				}
			}
		}
	},
}

var fetchCoursesCmd = &cobra.Command{
	Use:   "courses",
	Short: "Fetch list of student courses",
	Long:  `Fetch the list of all student courses from Central University.`,
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		client := mustClient()

		courses, err := client.GetStudentCourses(ctx, 10000, "published")
		if err != nil {
			fmt.Printf("Failed to fetch courses: %v\n", err)
			return
		}

		fmt.Printf("Found %d courses\n\n", len(courses.Items))
		for i, course := range courses.Items {
			fmt.Printf("%d. %s (ID: %d)\n", i+1, course.Name, course.ID)
			fmt.Printf("   State: %s | Archived: %v\n", course.State, course.IsArchived)

			if course.PublishedAt != nil {
				fmt.Printf("   Published: %s\n", course.PublishedAt.Format("2006-01-02 15:04:05"))
			}

			if course.Progress != nil {
				fmt.Printf("   Progress: %d/%d (%.1f%%)\n",
					course.Progress.CompletedCount,
					course.Progress.TotalCount,
					course.Progress.Percentage)
			}
			fmt.Println()
		}
	},
}

// clientForDownload creates a new client instance for concurrent downloads.
func clientForDownload() *cu.Client {
	client, _ := cu.NewClientFromEnv()
	return client
}
