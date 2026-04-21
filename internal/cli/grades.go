package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var gradesCmd = &cobra.Command{
	Use:   "grades [course]",
	Short: "Show grades and performance",
	Long: `Show grades for a specific course or a summary across all courses.
Course can be specified by ID or by name (substring match).

Examples:
  cu grades                 # summary across all courses
  cu grades go              # detailed grades for the Go course
  cu grades 901             # detailed grades for course 901`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		if len(args) == 0 {
			// Summary across all courses.
			courses, err := client.GetStudentCourses(ctx, 10000, "published")
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to fetch courses: %v\n", err)
				return
			}

			fmt.Println("Grades summary\n")
			for _, course := range courses.Items {
				progress, err := client.GetCourseProgress(ctx, course.ID)
				if err != nil {
					fmt.Printf("  %-50s  (error)\n", course.Name)
					continue
				}
				bar := progressBar(progress.EarnedScore, progress.MaxScore, 20)
				fmt.Printf("  %-50s  %s %.1f/%.0f\n",
					truncate(course.Name, 50),
					bar,
					progress.EarnedScore,
					progress.MaxScore,
				)
			}
			fmt.Println("\nUse: cu grades <course> for detailed view")
			return
		}

		// Detailed grades for a specific course.
		courseID, courseName := mustResolveCourse(ctx, client, args[0])
		fmt.Printf("Grades: %s\n\n", courseName)

		// Fetch activities performance (weighted breakdown).
		ap, err := client.GetActivitiesPerformance(ctx, courseID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch performance: %v\n", err)
			return
		}

		fmt.Println("Activity breakdown:")
		for _, item := range ap.Items {
			blocker := ""
			if item.IsBlocker {
				blocker = " [BLOCKER]"
			}
			weight := ""
			if item.Activity.Weight > 0 {
				weight = fmt.Sprintf(" (%.0f%%)", item.Activity.Weight*100)
			}
			fmt.Printf("  %-35s  avg=%.1f  total=%.1f%s%s\n",
				item.Activity.Name+weight,
				item.Average,
				item.Total,
				blocker,
				"",
			)
		}
		fmt.Printf("\n  Total score: %.1f\n\n", ap.TotalScore)

		// Fetch per-exercise scores.
		sp, err := client.GetStudentPerformance(ctx, courseID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch scores: %v\n", err)
			return
		}

		// Fetch exercises to get names.
		exercises, err := client.GetCourseExercises(ctx, courseID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch exercises: %v\n", err)
			return
		}
		nameByExerciseID := make(map[int]string)
		for _, ex := range exercises.Exercises {
			nameByExerciseID[ex.ID] = ex.Name
		}

		fmt.Println("Tasks:")
		for _, task := range sp.Tasks {
			name := nameByExerciseID[task.ExerciseID]
			if name == "" {
				name = fmt.Sprintf("exercise#%d", task.ExerciseID)
			}
			score := "  -"
			if task.Score != nil {
				score = fmt.Sprintf("%3.0f", *task.Score)
			}
			fmt.Printf("  %-12s  %s/%d  %s\n",
				stateLabel(task.State),
				score,
				task.MaxScore,
				name,
			)
		}

		if len(sp.Blockers) > 0 {
			fmt.Println("\nBlockers:")
			for _, b := range sp.Blockers {
				fmt.Printf("  %s (need avg >= %.0f)\n", b.ActivityName, b.AverageScoreThreshold)
			}
		}
	},
}

func progressBar(value, max float64, width int) string {
	if max == 0 {
		return "[" + repeat("-", width) + "]"
	}
	filled := int(value / max * float64(width))
	if filled > width {
		filled = width
	}
	return "[" + repeat("#", filled) + repeat("-", width-filled) + "]"
}

func repeat(s string, n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n-1]) + "…"
}
