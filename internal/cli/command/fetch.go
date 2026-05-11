package command

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"cu-sync/internal/model"
	"cu-sync/internal/render"
	"cu-sync/internal/usecase/materials"

	"github.com/spf13/cobra"
)

const maxCoursesLimit = 10000

func init() {
	fetchCourseCmd.Flags().String("path", ".", "path to save the course data")
	fetchCourseCmd.Flags().Bool("dump", false, "dumps all course data")

	fetchCmd.AddCommand(fetchCourseCmd)
	fetchCmd.AddCommand(fetchCourseSummaryCmd)
	fetchCmd.AddCommand(fetchCoursesCmd)
	fetchCmd.AddCommand(fetchStudentCmd)
	fetchCmd.AddCommand(fetchThemeCmd)
	fetchCmd.AddCommand(fetchLongreadCmd)
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
			exitErrf("Invalid course ID %q: %v", args[0], err)
		}

		fmt.Printf("Fetching course %d...\n", courseID)
		course, err := client.GetCourseOverview(ctx, courseID)
		if err != nil {
			exitErrf("Failed to fetch course: %v", err)
		}

		fmt.Printf("Course: %s (ID: %d)\n", course.Name, course.ID)
		fmt.Printf("State: %s | Archived: %v\n", course.State, course.IsArchived)
		fmt.Printf("Themes: %d\n\n", len(course.Themes))

		dump, _ := cmd.Flags().GetBool("dump")

		if dump {
			basePath, _ := cmd.Flags().GetString("path")
			courseDir := filepath.Join(basePath, materials.SanitizeFilename(course.Name)+strconv.Itoa(courseID))

			fmt.Println("Downloading course materials...")

			uc := materials.New(client, nil)
			onEvent := func(event model.MaterialEvent) {
				switch event.Type {
				case model.MaterialEventSaved:
					fmt.Printf("  saved: %s\n", event.Message)
				case model.MaterialEventError:
					fmt.Fprintf(os.Stderr, "  %s\n", event.Message)
				case model.MaterialEventTheme, model.MaterialEventPDF, model.MaterialEventLink:
					// skip verbose output in dump mode
				}
			}

			result, err := uc.Download(ctx, model.MaterialsDownloadInput{
				CourseQuery: args[0],
				BasePath:    courseDir,
			}, onEvent)
			if err != nil {
				exitErrf("Failed to download course data: %v", err)
			}

			fmt.Printf("Download complete: %d/%d files to %s\n",
				result.DownloadedFiles, result.TotalFiles, courseDir)
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

		courses, err := client.GetStudentCourses(ctx, maxCoursesLimit, "published")
		if err != nil {
			exitErrf("Failed to fetch courses: %v", err)
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

var fetchCourseSummaryCmd = &cobra.Command{
	Use:   "course-summary [course-id]",
	Short: "Fetch course summary by ID",
	Long:  `Fetch course summary information from Central University by course ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		courseID, err := strconv.Atoi(args[0])
		if err != nil {
			exitErrf("Invalid course ID %q: %v", args[0], err)
		}

		course, err := client.GetCourse(ctx, courseID)
		if err != nil {
			exitErrf("Failed to fetch course summary: %v", err)
		}

		fmt.Printf("Course: %s (ID: %d)\n", course.Name, course.ID)
		fmt.Printf("State: %s | Archived: %v\n", course.State, course.IsArchived)
		if course.Category != "" {
			fmt.Printf("Category: %s\n", course.Category)
		}
		if course.CategoryCover != "" {
			fmt.Printf("Category cover: %s\n", course.CategoryCover)
		}
		if course.SubjectID != nil {
			fmt.Printf("Subject ID: %d\n", *course.SubjectID)
		}
		fmt.Printf("Skill level: %s (enabled: %t)\n",
			course.Settings.SkillLevel, course.Settings.IsSkillLevelEnabled)
		render.TimeLine(os.Stdout, "Publish date", course.PublishDate)
		render.TimeLine(os.Stdout, "Published at", course.PublishedAt)
	},
}

var fetchStudentCmd = &cobra.Command{
	Use:   "student",
	Short: "Fetch current student profile",
	Long:  "Fetch the current student profile from Central University.",
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := cmd.Context()
		client := mustClient()

		student, err := client.GetCurrentStudent(ctx)
		if err != nil {
			exitErrf("Failed to fetch student profile: %v", err)
		}

		fullName := strings.TrimSpace(strings.Join(
			[]string{student.LastName, student.FirstName, student.MiddleName}, " "))
		if fullName == "" {
			fullName = "Unknown"
		}

		fmt.Printf("Student: %s\n", fullName)
		fmt.Printf("ID: %s\n", student.ID)
		if student.UniversityEmail != "" {
			fmt.Printf("Email: %s\n", student.UniversityEmail)
		}
		if student.StudyLevel != "" {
			fmt.Printf("Study level: %s\n", student.StudyLevel)
		}
		if student.StudyStartYear > 0 {
			fmt.Printf("Study start year: %d\n", student.StudyStartYear)
		}
		if student.TimeAccount != "" {
			fmt.Printf("Time account: %s\n", student.TimeAccount)
		}
		fmt.Printf("Late days balance: %d\n", student.LateDaysBalance)
	},
}

var fetchThemeCmd = &cobra.Command{
	Use:   "theme [theme-id]",
	Short: "Fetch theme summary by ID",
	Long:  `Fetch theme summary information from Central University by theme ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		themeID, err := strconv.Atoi(args[0])
		if err != nil {
			exitErrf("Invalid theme ID %q: %v", args[0], err)
		}

		theme, err := client.GetTheme(ctx, themeID)
		if err != nil {
			exitErrf("Failed to fetch theme: %v", err)
		}

		fmt.Printf("Theme: %s (ID: %d)\n", theme.Name, theme.ID)
		fmt.Printf("State: %s | Order: %d\n", theme.State, theme.Order)
		render.TimeLine(os.Stdout, "Publish date", theme.PublishDate)
		render.TimeLine(os.Stdout, "Published at", theme.PublishedAt)
	},
}

var fetchLongreadCmd = &cobra.Command{
	Use:   "longread [longread-id]",
	Short: "Fetch longread summary by ID",
	Long:  `Fetch longread summary information from Central University by longread ID.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		longreadID, err := strconv.Atoi(args[0])
		if err != nil {
			exitErrf("Invalid longread ID %q: %v", args[0], err)
		}

		longread, err := client.GetLongread(ctx, longreadID)
		if err != nil {
			exitErrf("Failed to fetch longread: %v", err)
		}

		fmt.Printf("Longread: %s (ID: %d)\n", longread.Name, longread.ID)
		fmt.Printf("Type: %s | State: %s | Order: %d\n", longread.Type, longread.State, longread.Order)
		render.TimeLine(os.Stdout, "Publish date", longread.PublishDate)
		render.TimeLine(os.Stdout, "Published at", longread.PublishedAt)
	},
}
