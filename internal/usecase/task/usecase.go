package task

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"cu-sync/internal/usecase/task/model/input"
	"cu-sync/internal/usecase/task/model/output"
)

const (
	hoursPerDay    = 24
	minutesPerHour = 60
	percentMul     = 100
)

// UseCase implements the task business logic.
type UseCase struct {
	lms LMSClient
}

// New creates a new task usecase.
func New(lms LMSClient) *UseCase {
	return &UseCase{lms: lms}
}

// Get fetches a task and returns it with computed fields.
func (uc *UseCase) Get(ctx context.Context, in input.GetInput) (*output.TaskOutput, error) {
	t, err := uc.lms.GetTask(ctx, in.TaskID)
	if err != nil {
		return nil, fmt.Errorf("fetching task %d: %w", in.TaskID, err)
	}

	out := &output.TaskOutput{
		CourseName:      t.Course.Name,
		ThemeName:       t.Theme.Name,
		ExerciseName:    t.Exercise.Name,
		ActivityName:    t.Exercise.Activity.Name,
		ActivityWeight:  t.Exercise.Activity.Weight * percentMul,
		Deadline:        t.Deadline,
		StartedAt:       t.StartedAt,
		SubmitAt:        t.SubmitAt,
		RejectAt:        t.RejectAt,
		EvaluateAt:      t.EvaluateAt,
		MaxScore:        t.Exercise.MaxScore,
		LateDaysBalance: t.Student.LateDaysBalance,
		StateLabel:      stateLabel(t.State),
		TimeLeft:        formatTimeLeft(t.Deadline),
	}

	if t.Score != nil {
		out.ScoreFormatted = fmt.Sprintf("%.0f/%d", *t.Score, t.Exercise.MaxScore)
	} else {
		out.ScoreFormatted = fmt.Sprintf("-/%d", t.Exercise.MaxScore)
	}

	if t.Reviewer != nil {
		out.ReviewerName = t.Reviewer.FirstName + " " + t.Reviewer.LastName
		out.ReviewerEmail = t.Reviewer.Email
	}

	if t.Solution != nil {
		out.SolutionURL = t.Solution.SolutionURL
	}

	return out, nil
}

func stateLabel(state string) string {
	switch state {
	case "backlog":
		return "TODO"
	case "inProgress":
		return "IN PROGRESS"
	case "submitted":
		return "SUBMITTED"
	case "evaluated":
		return "DONE"
	case "failed":
		return "FAILED"
	default:
		return strings.ToUpper(state)
	}
}

func formatTimeLeft(t time.Time) string {
	d := time.Until(t)
	if d < 0 {
		return "OVERDUE"
	}
	days := int(d.Hours() / hoursPerDay)
	hours := int(math.Mod(d.Hours(), hoursPerDay))
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	if hours > 0 {
		return fmt.Sprintf("%dh %dm", hours, int(math.Mod(d.Minutes(), minutesPerHour)))
	}
	return fmt.Sprintf("%dm", int(d.Minutes()))
}
