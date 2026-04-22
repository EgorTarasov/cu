package command

import (
	"fmt"

	"cu-sync/internal/model"
	"cu-sync/internal/usecase/deadlines"

	"github.com/spf13/cobra"
)

var deadlinesCmd = &cobra.Command{
	Use:   "deadlines [course]",
	Short: "Show upcoming deadlines",
	Long: `Show upcoming deadlines across all courses or for a specific course.
Course can be specified by ID or by name (substring match).

Examples:
  cu deadlines              # all upcoming deadlines
  cu deadlines go           # deadlines for the Go course
  cu deadlines 901          # deadlines for course ID 901
  cu deadlines алгоритмы    # deadlines matching "алгоритмы"`,
	Args: cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		in := model.DeadlinesListInput{}
		if len(args) > 0 {
			in.CourseQuery = args[0]
		}

		uc := deadlines.New(client)
		result, err := uc.List(ctx, in)
		if err != nil {
			fmt.Printf("Failed to fetch deadlines: %v\n", err)
			return
		}

		printDeadlines(result)
	},
}

func printDeadlines(result *model.DeadlinesListOutput) {
	if len(result.Items) == 0 {
		fmt.Println("No upcoming deadlines!")
		return
	}

	if result.CourseName != "" {
		fmt.Printf("Deadlines: %s\n\n", result.CourseName)
	} else {
		fmt.Println("All upcoming deadlines")
		fmt.Println()
	}

	for _, dl := range result.Items {
		marker := " "
		switch dl.Deadline.Urgency() {
		case model.UrgencyUrgent:
			marker = "!"
		case model.UrgencySoon:
			marker = "*"
		case model.UrgencyNormal:
			// no marker
		}

		fmt.Printf(" %s %-12s  %-8s  %s  %s\n",
			marker,
			dl.StateLabel,
			dl.Deadline.TimeLeft(),
			dl.Deadline.Format(model.DateTimeShortFormat),
			dl.ExerciseName,
		)

		if result.CourseName == "" {
			fmt.Printf("   %s\n", dl.CourseName)
		}

		if dl.Reviewer != nil {
			fmt.Printf("   reviewer: %s\n", dl.Reviewer.FullName())
		}
	}

	fmt.Printf("\n%d deadline(s) total\n", len(result.Items))
	fmt.Println("  ! = overdue or <24h  * = <3 days")
}
