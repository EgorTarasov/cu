package cli

import (
	"fmt"
	"os"

	"cu-sync/internal/format"
	"cu-sync/internal/usecase/grades"
	"cu-sync/internal/usecase/grades/model/input"
	"cu-sync/internal/usecase/grades/model/output"

	"github.com/spf13/cobra"
)

const (
	gradeNameWidth      = 50
	gradeProgressBarLen = 20
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
		uc := grades.New(client)

		if len(args) == 0 {
			result, err := uc.Summary(ctx, input.SummaryInput{})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to fetch grades: %v\n", err)
				return
			}
			printSummary(result)
			return
		}

		result, err := uc.Detailed(ctx, input.DetailedInput{CourseQuery: args[0]})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch grades: %v\n", err)
			return
		}
		printDetailed(result)
	},
}

func printSummary(result *output.SummaryOutput) {
	fmt.Println("Grades summary")
	fmt.Println()
	for _, item := range result.Items {
		if item.Error != nil {
			fmt.Printf("  %-50s  (error)\n", item.CourseName)
			continue
		}
		bar := format.ProgressBar(item.EarnedScore, item.MaxScore, gradeProgressBarLen)
		fmt.Printf("  %-50s  %s %.1f/%.0f\n",
			format.Truncate(item.CourseName, gradeNameWidth),
			bar,
			item.EarnedScore,
			item.MaxScore,
		)
	}
	fmt.Println("\nUse: cu grades <course> for detailed view")
}

func printDetailed(result *output.DetailedOutput) {
	fmt.Printf("Grades: %s\n\n", result.CourseName)

	fmt.Println("Activity breakdown:")
	for _, item := range result.Activities {
		blocker := ""
		if item.IsBlocker {
			blocker = " [BLOCKER]"
		}
		weight := ""
		if item.Weight > 0 {
			weight = fmt.Sprintf(" (%.0f%%)", item.Weight)
		}
		fmt.Printf("  %-35s  avg=%.1f  total=%.1f%s%s\n",
			item.Name+weight,
			item.Average,
			item.Total,
			blocker,
			"",
		)
	}
	fmt.Printf("\n  Total score: %.1f\n\n", result.TotalScore)

	fmt.Println("Tasks:")
	for _, t := range result.Tasks {
		score := "  -"
		if t.Score != nil {
			score = fmt.Sprintf("%3.0f", *t.Score)
		}
		fmt.Printf("  %-12s  %s/%d  %s\n",
			t.StateLabel,
			score,
			t.MaxScore,
			t.Name,
		)
	}

	if len(result.Blockers) > 0 {
		fmt.Println("\nBlockers:")
		for _, b := range result.Blockers {
			fmt.Printf("  %s (need avg >= %.0f)\n", b.ActivityName, b.Threshold)
		}
	}
}
