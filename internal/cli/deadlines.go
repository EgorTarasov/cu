package cli

import (
	"fmt"
	"sort"
	"time"

	"github.com/spf13/cobra"
)

const deadlinesLimit = 100

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

		var courseID *int
		var courseName string

		if len(args) > 0 {
			id, name := mustResolveCourse(ctx, client, args[0])
			courseID = &id
			courseName = name
		}

		deadlines, err := client.GetDeadlines(ctx, deadlinesLimit, courseID)
		if err != nil {
			fmt.Printf("Failed to fetch deadlines: %v\n", err)
			return
		}

		if len(deadlines) == 0 {
			fmt.Println("No upcoming deadlines!")
			return
		}

		// Sort by deadline date.
		sort.Slice(deadlines, func(i, j int) bool {
			return deadlines[i].Deadline.Before(deadlines[j].Deadline)
		})

		if courseName != "" {
			fmt.Printf("Deadlines: %s\n\n", courseName)
		} else {
			fmt.Println("All upcoming deadlines")
			fmt.Println()
		}

		now := time.Now()
		for _, dl := range deadlines {
			timeLeft := formatTimeLeft(dl.Deadline)
			state := stateLabel(dl.State)

			// Urgency marker.
			marker := " "
			remaining := dl.Deadline.Sub(now)
			switch {
			case remaining < 0:
				marker = "!"
			case remaining < 24*time.Hour:
				marker = "!"
			case remaining < 3*24*time.Hour:
				marker = "*"
			}

			fmt.Printf(" %s %-12s  %-8s  %s  %s\n",
				marker,
				state,
				timeLeft,
				dl.Deadline.Format("02 Jan 15:04"),
				dl.Exercise.Name,
			)

			if courseID == nil {
				fmt.Printf("   %s\n", dl.Course.Name)
			}

			if dl.Reviewer != nil {
				fmt.Printf("   reviewer: %s %s\n", dl.Reviewer.FirstName, dl.Reviewer.LastName)
			}
		}

		fmt.Printf("\n%d deadline(s) total\n", len(deadlines))
		fmt.Println("  ! = overdue or <24h  * = <3 days")
	},
}
