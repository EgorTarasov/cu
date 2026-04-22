package courses

import (
	"context"
	"fmt"

	"cu-sync/internal/model"
)

const maxCoursesLimit = 10000

type UseCase struct {
	lms LMSClient
}

func New(lms LMSClient) *UseCase {
	return &UseCase{lms: lms}
}

func (uc *UseCase) List(ctx context.Context) (*model.CoursesListOutput, error) {
	courses, err := uc.lms.GetStudentCourses(ctx, maxCoursesLimit, "published")
	if err != nil {
		return nil, fmt.Errorf("fetching courses: %w", err)
	}

	items := make([]model.CourseItem, 0, len(courses.Items))
	for _, c := range courses.Items {
		items = append(items, model.CourseItem{
			ID:   c.ID,
			Name: c.Name,
		})
	}

	return &model.CoursesListOutput{Items: items}, nil
}
