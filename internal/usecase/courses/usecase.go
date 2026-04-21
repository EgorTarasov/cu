package courses

import (
	"context"
	"fmt"

	"cu-sync/internal/usecase/courses/model/output"
)

const maxCoursesLimit = 10000

// UseCase implements the courses business logic.
type UseCase struct {
	lms LMSClient
}

// New creates a new courses usecase.
func New(lms LMSClient) *UseCase {
	return &UseCase{lms: lms}
}

// List fetches all published courses for the student.
func (uc *UseCase) List(ctx context.Context) (*output.ListOutput, error) {
	courses, err := uc.lms.GetStudentCourses(ctx, maxCoursesLimit, "published")
	if err != nil {
		return nil, fmt.Errorf("fetching courses: %w", err)
	}

	items := make([]output.CourseItem, 0, len(courses.Items))
	for _, c := range courses.Items {
		items = append(items, output.CourseItem{
			ID:   c.ID,
			Name: c.Name,
		})
	}

	return &output.ListOutput{Items: items}, nil
}
