package grades

import (
	"context"
	"fmt"
	"strings"

	"cu-sync/internal/usecase/grades/model/input"
	"cu-sync/internal/usecase/grades/model/output"
)

const (
	maxCoursesLimit   = 10000
	percentMultiplier = 100
)

// UseCase implements the grades business logic.
type UseCase struct {
	lms LMSClient
}

// New creates a new grades usecase.
func New(lms LMSClient) *UseCase {
	return &UseCase{lms: lms}
}

// Summary returns a grades summary across all published courses.
func (uc *UseCase) Summary(ctx context.Context, _ input.SummaryInput) (*output.SummaryOutput, error) {
	courses, err := uc.lms.GetStudentCourses(ctx, maxCoursesLimit, "published")
	if err != nil {
		return nil, fmt.Errorf("fetching courses: %w", err)
	}

	items := make([]output.SummaryItem, 0, len(courses.Items))
	for _, course := range courses.Items {
		item := output.SummaryItem{CourseName: course.Name}

		progress, err := uc.lms.GetCourseProgress(ctx, course.ID)
		if err != nil {
			item.Error = err
		} else {
			item.EarnedScore = progress.EarnedScore
			item.MaxScore = progress.MaxScore
		}

		items = append(items, item)
	}

	return &output.SummaryOutput{Items: items}, nil
}

// Detailed returns detailed grades for a specific course.
func (uc *UseCase) Detailed(ctx context.Context, in input.DetailedInput) (*output.DetailedOutput, error) {
	courseID, courseName, err := uc.lms.ResolveCourse(ctx, in.CourseQuery)
	if err != nil {
		return nil, fmt.Errorf("resolving course: %w", err)
	}

	// Fetch activities performance (weighted breakdown).
	ap, err := uc.lms.GetActivitiesPerformance(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("fetching activities performance: %w", err)
	}

	activities := make([]output.ActivityBreakdown, 0, len(ap.Items))
	for _, item := range ap.Items {
		activities = append(activities, output.ActivityBreakdown{
			Name:      item.Activity.Name,
			Weight:    item.Activity.Weight * percentMultiplier,
			Average:   item.Average,
			Total:     item.Total,
			IsBlocker: item.IsBlocker,
		})
	}

	// Fetch per-exercise scores.
	sp, err := uc.lms.GetStudentPerformance(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("fetching student performance: %w", err)
	}

	// Fetch exercises to build name map.
	exercises, err := uc.lms.GetCourseExercises(ctx, courseID)
	if err != nil {
		return nil, fmt.Errorf("fetching course exercises: %w", err)
	}

	nameByExerciseID := make(map[int]string, len(exercises.Exercises))
	for _, ex := range exercises.Exercises {
		nameByExerciseID[ex.ID] = ex.Name
	}

	tasks := make([]output.TaskGrade, 0, len(sp.Tasks))
	for _, t := range sp.Tasks {
		name := nameByExerciseID[t.ExerciseID]
		if name == "" {
			name = fmt.Sprintf("exercise#%d", t.ExerciseID)
		}

		tasks = append(tasks, output.TaskGrade{
			Name:       name,
			State:      t.State,
			StateLabel: stateLabel(t.State),
			Score:      t.Score,
			MaxScore:   t.MaxScore,
		})
	}

	blockers := make([]output.BlockerInfo, 0, len(sp.Blockers))
	for _, b := range sp.Blockers {
		blockers = append(blockers, output.BlockerInfo{
			ActivityName: b.ActivityName,
			Threshold:    b.AverageScoreThreshold,
		})
	}

	return &output.DetailedOutput{
		CourseName: courseName,
		Activities: activities,
		TotalScore: ap.TotalScore,
		Tasks:      tasks,
		Blockers:   blockers,
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
