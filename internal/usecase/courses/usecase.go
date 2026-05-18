package courses

import (
	"context"
	"fmt"

	"github.com/EgorTarasov/cu/internal/gateway/cu"
	"github.com/EgorTarasov/cu/internal/model"
)

const maxCoursesLimit = 10000

type UseCase struct {
	lms LMSClient
}

func New(lms LMSClient) *UseCase {
	return &UseCase{lms: lms}
}

func (uc *UseCase) List(ctx context.Context) (*model.CoursesListOutput, error) {
	active, err := uc.lms.GetStudentCourses(ctx, maxCoursesLimit, "published")
	if err != nil {
		return nil, fmt.Errorf("fetching active courses: %w", err)
	}
	archived, err := uc.lms.GetStudentCourses(ctx, maxCoursesLimit, "archived")
	if err != nil {
		return nil, fmt.Errorf("fetching archived courses: %w", err)
	}

	return &model.CoursesListOutput{
		Active:   toCourseItems(active.Items, false),
		Archived: toCourseItems(archived.Items, true),
	}, nil
}

func toCourseItems(in []cu.StudentCourse, archived bool) []model.CourseItem {
	out := make([]model.CourseItem, 0, len(in))
	for _, c := range in {
		out = append(out, model.CourseItem{
			ID:         c.ID,
			Name:       c.Name,
			IsArchived: archived,
		})
	}
	return out
}
