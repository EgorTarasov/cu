package cli

import (
	"cu-sync/internal/cu"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

func init() {
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
		fmt.Println("📚 Fetching Course Overview")
		fmt.Println("===========================")
		fmt.Println()

		courseID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid course ID '%s': %v", args[0], err)
		}

		bffCookie := os.Getenv("CU_BFF_COOKIE")
		if bffCookie == "" {
			fmt.Println("⚠️  No CU_BFF_COOKIE environment variable found.")
			fmt.Println("Please set the CU_BFF_COOKIE environment variable with your bff.cookie value:")
			fmt.Println()
			fmt.Println("Example:")
			fmt.Println("  export CU_BFF_COOKIE='your-cookie-value-here'")
			fmt.Println("  cu fetch course 519")
			fmt.Println()
			log.Fatal("CU_BFF_COOKIE environment variable is required")
		}

		client := cu.NewClient(bffCookie)

		if err := client.ValidateCookie(); err != nil {
			fmt.Printf("⚠️  Cookie validation failed: %v\n", err)
			fmt.Println("The stored cookie might be expired. Please update it.")
			return
		}

		fmt.Printf("Fetching course %d...\n", courseID)
		course, err := client.GetCourseOverview(courseID)
		if err != nil {
			log.Fatalf("Failed to fetch course: %v", err)
		}

		fmt.Println("✅ Course fetched successfully!")
		fmt.Println()
		fmt.Printf("📖 Course: %s (ID: %d)\n", course.Name, course.ID)
		fmt.Printf("📊 State: %s\n", course.State)
		fmt.Printf("📁 Archived: %v\n", course.IsArchived)

		if course.PublishDate != nil {
			fmt.Printf("📅 Publish Date: %s\n", course.PublishDate.Format("2006-01-02 15:04:05"))
		}

		fmt.Printf("🎯 Skill Level: %s (Enabled: %v)\n",
			course.Settings.SkillLevel,
			course.Settings.IsSkillLevelEnabled)

		fmt.Printf("📚 Themes: %d\n", len(course.Themes))

		for i, theme := range course.Themes {
			fmt.Printf("  %d. %s (ID: %d)\n", theme.Order, theme.Name, theme.ID)
			fmt.Printf("     📖 Longreads: %d\n", len(theme.Longreads))

			for _, longread := range theme.Longreads {
				fmt.Printf("       - %s (%s)\n", longread.Name, longread.Type)
				if len(longread.Exercises) > 0 {
					fmt.Printf("         📝 Exercises: %d\n", len(longread.Exercises))
				}
			}

			if i < len(course.Themes)-1 {
				fmt.Println()
			}
		}

		fmt.Println()
		fmt.Println("💡 Course data fetched successfully using CU_BFF_COOKIE environment variable.")
	},
}

var fetchCoursesCmd = &cobra.Command{
	Use:   "courses",
	Short: "Fetch list of student courses",
	Long:  `Fetch the list of all student courses from Central University.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📚 Fetching Student Courses")
		fmt.Println("===========================")
		fmt.Println()

		bffCookie := os.Getenv("CU_BFF_COOKIE")
		if bffCookie == "" {
			fmt.Println("⚠️  No CU_BFF_COOKIE environment variable found.")
			fmt.Println("Please set the CU_BFF_COOKIE environment variable with your bff.cookie value:")
			fmt.Println()
			fmt.Println("Example:")
			fmt.Println("  export CU_BFF_COOKIE='your-cookie-value-here'")
			fmt.Println("  cu fetch courses")
			fmt.Println()
			log.Fatal("CU_BFF_COOKIE environment variable is required")
		}

		client := cu.NewClient(bffCookie)

		if err := client.ValidateCookie(); err != nil {
			fmt.Printf("⚠️  Cookie validation failed: %v\n", err)
			fmt.Println("The CU_BFF_COOKIE might be expired. Please update it.")
			return
		}

		fmt.Println("Fetching all published courses...")
		courses, err := client.GetStudentCourses(10000, "published")
		if err != nil {
			log.Fatalf("Failed to fetch courses: %v", err)
		}

		fmt.Printf("✅ Successfully fetched %d courses!\n", len(courses.Items))
		fmt.Printf("📊 Total available: %d courses\n", courses.Paging.TotalCount)
		fmt.Println()

		for i, course := range courses.Items {
			fmt.Printf("%d. 📖 %s (ID: %d)\n", i+1, course.Name, course.ID)
			fmt.Printf("   📊 State: %s | 📁 Archived: %v\n", course.State, course.IsArchived)

			if course.PublishedAt != nil {
				fmt.Printf("   📅 Published: %s\n", course.PublishedAt.Format("2006-01-02 15:04:05"))
			}

			fmt.Printf("   🎯 Skill Level: %s (Enabled: %v)\n",
				course.Settings.SkillLevel,
				course.Settings.IsSkillLevelEnabled)

			if course.Progress != nil {
				fmt.Printf("   📈 Progress: %d/%d (%.1f%%)\n",
					course.Progress.CompletedCount,
					course.Progress.TotalCount,
					course.Progress.Percentage)
			}

			fmt.Println()
		}

		fmt.Printf("💡 Found %d published courses total.\n", len(courses.Items))
		fmt.Println("Use 'cu fetch course [ID]' to get detailed information about a specific course.")
	},
}
