package command

import (
	"fmt"
	"os"
	"strconv"

	"cu-sync/internal/model"
	"cu-sync/internal/usecase/task"

	"github.com/spf13/cobra"
)

var taskCmd = &cobra.Command{
	Use:   "task <task-id>",
	Short: "Show task details",
	Long: `Show detailed information about a specific task (assignment instance).
Task IDs can be found via 'cu deadlines' or 'cu grades'.

Examples:
  cu task 1536681`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := cmd.Context()
		client := mustClient()

		taskID, err := strconv.Atoi(args[0])
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid task ID: %s\n", args[0])
			os.Exit(1)
		}

		uc := task.New(client)
		out, err := uc.Get(ctx, model.TaskGetInput{TaskID: taskID})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch task: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Task: %s\n", out.ExerciseName)
		fmt.Printf("Course: %s\n", out.CourseName)
		fmt.Printf("Theme: %s\n", out.ThemeName)
		fmt.Println()

		fmt.Printf("State:    %s\n", out.StateLabel)
		fmt.Printf("Score:    %s\n", out.ScoreFormatted)
		fmt.Printf("Activity: %s (%.0f%%)\n", out.ActivityName, out.ActivityWeight)
		fmt.Println()

		fmt.Printf("Deadline: %s (%s left)\n",
			out.Deadline.Format("02 Jan 2006 15:04"),
			out.TimeLeft,
		)
		if out.StartedAt != nil {
			fmt.Printf("Started:  %s\n", out.StartedAt.Format("02 Jan 2006 15:04"))
		}
		if out.SubmitAt != nil {
			fmt.Printf("Submitted: %s\n", out.SubmitAt.Format("02 Jan 2006 15:04"))
		}
		if out.RejectAt != nil {
			fmt.Printf("Rejected: %s\n", out.RejectAt.Format("02 Jan 2006 15:04"))
		}
		if out.EvaluateAt != nil {
			fmt.Printf("Evaluated: %s\n", out.EvaluateAt.Format("02 Jan 2006 15:04"))
		}
		fmt.Println()

		if out.ReviewerName != "" {
			fmt.Printf("Reviewer: %s (%s)\n", out.ReviewerName, out.ReviewerEmail)
		}

		if out.SolutionURL != "" {
			fmt.Printf("Solution: %s\n", out.SolutionURL)
		}

		fmt.Printf("\nLate days balance: %d\n", out.LateDaysBalance)
	},
}
