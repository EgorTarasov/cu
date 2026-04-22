package deadlines

import (
	"context"
	"fmt"
	"sort"

	"cu-sync/internal/model"
)

const deadlinesLimit = 100

type UseCase struct {
	lms LMSClient
}

func New(lms LMSClient) *UseCase {
	return &UseCase{lms: lms}
}

func (uc *UseCase) List(ctx context.Context, in model.DeadlinesListInput) (*model.DeadlinesListOutput, error) {
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

	sort.Slice(deadlines, func(i, j int) bool {
		return deadlines[i].Deadline.Before(deadlines[j].Deadline)
	})

	items := make([]model.DeadlineItem, 0, len(deadlines))

	for _, dl := range deadlines {
		item := model.DeadlineItem{
			ExerciseName: dl.Exercise.Name,
			CourseName:   dl.Course.Name,
			State:        model.TaskState(dl.State),
			Deadline:     model.DeadLine(dl.Deadline),
		}
		if dl.Reviewer != nil {
			item.Reviewer = &model.Reviewer{
				FirstName: dl.Reviewer.FirstName,
				LastName:  dl.Reviewer.LastName,
				Email:     dl.Reviewer.Email,
			}
		}
		items = append(items, item)
	}

	return &model.DeadlinesListOutput{
		Items:      items,
		CourseName: courseName,
	}, nil
}
