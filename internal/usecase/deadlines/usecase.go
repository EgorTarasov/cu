package deadlines

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"cu-sync/internal/usecase/deadlines/model/input"
	"cu-sync/internal/usecase/deadlines/model/output"
)

const (
	deadlinesLimit = 100
	hoursPerDay    = 24
	minutesPerHour = 60
)

// UseCase implements the deadlines business logic.
type UseCase struct {
	lms LMSClient
}

// New creates a new deadlines usecase.
func New(lms LMSClient) *UseCase {
	return &UseCase{lms: lms}
}

// List fetches deadlines, sorts by date, and classifies urgency.
func (uc *UseCase) List(ctx context.Context, in input.ListInput) (*output.ListOutput, error) {
	var courseID *int
	var courseName string

	if in.CourseQuery != "" {
		id, name, err := uc.lms.ResolveCourse(ctx, in.CourseQuery)
		if err != nil {
			return nil, fmt.Errorf("resolving course: %w", err)
		}
		courseID = &id
		courseName = name
	}

	deadlines, err := uc.lms.GetDeadlines(ctx, deadlinesLimit, courseID)
	if err != nil {
		return nil, fmt.Errorf("fetching deadlines: %w", err)
	}

	// Sort by deadline ascending.
	sort.Slice(deadlines, func(i, j int) bool {
		return deadlines[i].Deadline.Before(deadlines[j].Deadline)
	})

	now := time.Now()
	items := make([]output.DeadlineItem, 0, len(deadlines))

	for _, dl := range deadlines {
		remaining := dl.Deadline.Sub(now)

		var urgency output.UrgencyLevel
		switch {
		case remaining < 0 || remaining < 24*time.Hour:
			urgency = output.UrgencyUrgent
		case remaining < 3*24*time.Hour:
			urgency = output.UrgencySoon
		default:
			urgency = output.UrgencyNormal
		}

		var reviewerName string
		if dl.Reviewer != nil {
			reviewerName = dl.Reviewer.FirstName + " " + dl.Reviewer.LastName
		}

		items = append(items, output.DeadlineItem{
			ExerciseName: dl.Exercise.Name,
			CourseName:   dl.Course.Name,
			State:        dl.State,
			StateLabel:   stateLabel(dl.State),
			Deadline:     dl.Deadline,
			TimeLeft:     formatTimeLeft(dl.Deadline),
			Urgency:      urgency,
			ReviewerName: reviewerName,
		})
	}

	return &output.ListOutput{
		Items:      items,
		CourseName: courseName,
	}, nil
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
