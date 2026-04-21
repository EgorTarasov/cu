package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/spf13/cobra"
)

const taskPercentMultiplier = 100

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

		task, err := client.GetTask(ctx, taskID)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to fetch task: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Task: %s\n", task.Exercise.Name)
		fmt.Printf("Course: %s\n", task.Course.Name)
		fmt.Printf("Theme: %s\n", task.Theme.Name)
		fmt.Println()

		fmt.Printf("State:    %s\n", stateLabel(task.State))
		if task.Score != nil {
			fmt.Printf("Score:    %.0f/%d\n", *task.Score, task.Exercise.MaxScore)
		} else {
			fmt.Printf("Score:    -/%d\n", task.Exercise.MaxScore)
		}
		fmt.Printf("Activity: %s (%.0f%%)\n",
			task.Exercise.Activity.Name,
			task.Exercise.Activity.Weight*taskPercentMultiplier)
		fmt.Println()

		fmt.Printf("Deadline: %s (%s left)\n",
			task.Deadline.Format("02 Jan 2006 15:04"),
			formatTimeLeft(task.Deadline),
		)
		if task.StartedAt != nil {
			fmt.Printf("Started:  %s\n", task.StartedAt.Format("02 Jan 2006 15:04"))
		}
		if task.SubmitAt != nil {
			fmt.Printf("Submitted: %s\n", task.SubmitAt.Format("02 Jan 2006 15:04"))
		}
		if task.RejectAt != nil {
			fmt.Printf("Rejected: %s\n", task.RejectAt.Format("02 Jan 2006 15:04"))
		}
		if task.EvaluateAt != nil {
			fmt.Printf("Evaluated: %s\n", task.EvaluateAt.Format("02 Jan 2006 15:04"))
		}
		fmt.Println()

		if task.Reviewer != nil {
			fmt.Printf("Reviewer: %s %s (%s)\n",
				task.Reviewer.FirstName, task.Reviewer.LastName, task.Reviewer.Email)
		}

		if task.Solution != nil && task.Solution.SolutionURL != "" {
			fmt.Printf("Solution: %s\n", task.Solution.SolutionURL)
		}

		fmt.Printf("\nLate days balance: %d\n", task.Student.LateDaysBalance)
	},
}
