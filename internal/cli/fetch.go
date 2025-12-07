package cli

import (
	"context"
	"cu-sync/internal/cu"
	"fmt"
	"log"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

const maxConcurrentDownloads = 10

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
		fmt.Println("📚 Fetching Course Overview")
		fmt.Println("===========================")
		fmt.Println()

		courseID, err := strconv.Atoi(args[0])
		if err != nil {
			log.Fatalf("Invalid course ID '%s': %v", args[0], err)
		}

		client, err := cu.NewClientFromEnv()
		if err != nil {
			cookieRequiredError(err)
		}

		if err := client.ValidateCookie(); err != nil {
			fmt.Printf("⚠️  Cookie validation failed: %v\n", err)
			fmt.Println("The stored cookie might be expired. Please update it.")
			return
		}

		fmt.Printf("Fetching course %d...\n", courseID)
		course, err := client.GetCourseOverview(ctx, courseID)
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

		dump, _ := cmd.Flags().GetBool("dump")

		if dump {
			basePath, _ := cmd.Flags().GetString("path")
			courseDir := filepath.Join(basePath, sanitizeFilename(course.Name)+strconv.Itoa(courseID))
			err = dumpCourse(ctx, client, course, courseDir)
			if err != nil {
				fmt.Printf("failed to download course data: %v\n", err)
				return
			}
		} else {
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
		}

		fmt.Println("💡 Course data fetched successfully using CU_BFF_COOKIE environment variable.")
	},
}

func sanitizeFilename(name string) string {
	replacements := map[rune]rune{
		'/':  '-',
		'\\': '-',
		':':  '-',
		'*':  '-',
		'?':  '-',
		'"':  '-',
		'<':  '-',
		'>':  '-',
		'|':  '-',
	}

	runes := []rune(name)
	for i, r := range runes {
		if replacement, ok := replacements[r]; ok {
			runes[i] = replacement
		}
	}

	return string(runes)
}

var fetchCoursesCmd = &cobra.Command{
	Use:   "courses",
	Short: "Fetch list of student courses",
	Long:  `Fetch the list of all student courses from Central University.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("📚 Fetching Student Courses")
		fmt.Println("===========================")
		fmt.Println()

		client, err := cu.NewClientFromEnv()
		if err != nil {
			cookieRequiredError(err)
		}

		if err := client.ValidateCookie(); err != nil {
			fmt.Printf("⚠️  Cookie validation failed: %v\n", err)
			fmt.Println("The CU_BFF_COOKIE might be expired. Please update it.")
			return
		}

		fmt.Println("Fetching all published courses...")
		courses, err := client.GetStudentCourses(cmd.Context(), 10000, "published")
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

func dumpCourse(
	ctx context.Context,
	client *cu.Client,
	course *cu.CourseOverview,
	courseDir string,
) error {
	fmt.Println()
	fmt.Println("📥 Downloading course materials...")
	fmt.Println()

	totalFiles := atomic.Int32{}
	downloadedFiles := atomic.Int32{}

	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(maxConcurrentDownloads)

	for _, theme := range course.Themes {
		theme := theme // capture loop variable
		themeDir := filepath.Join(courseDir, fmt.Sprintf("%d-%s", theme.Order, sanitizeFilename(theme.Name)))

		fmt.Printf("📁 Theme: %s\n", theme.Name)

		for _, longread := range theme.Longreads {
			longread := longread // capture loop variable
			longreadDir := filepath.Join(themeDir, sanitizeFilename(longread.Name))

			fmt.Printf("  📖 Longread: %s\n", longread.Name)

			materials, err := client.GetLongReadContent(gctx, longread.ID)
			if err != nil {
				fmt.Printf("    ⚠️  Failed to fetch materials: %v\n", err)
				continue
			}

			fileCount := 0
			for _, material := range materials.Items {
				if material.Discriminator == "file" {
					fileCount++
				}
			}

			if fileCount == 0 {
				fmt.Printf("    ℹ️  No files to download\n")
				continue
			}

			totalFiles.Add(int32(fileCount))

			for _, material := range materials.Items {
				if material.Discriminator == "file" {
					material := material

					g.Go(func() error {
						filePath, err := client.DownloadFile(gctx, material, longreadDir)
						if err != nil {
							fmt.Printf("    ❌ Failed to download %s: %v\n", material.Content.Name, err)
							return nil
						}

						downloadedFiles.Add(1)
						fmt.Printf("    ✅ Downloaded: %s\n", filepath.Base(filePath))
						return nil
					})
				}
			}
		}
		fmt.Println()
	}

	if err := g.Wait(); err != nil {
		return fmt.Errorf("error downloading course materials: %w", err)
	}

	fmt.Printf("✅ Download complete! %d/%d files downloaded to %s\n", downloadedFiles.Load(), totalFiles.Load(), courseDir)
	fmt.Println()
	return nil
}
