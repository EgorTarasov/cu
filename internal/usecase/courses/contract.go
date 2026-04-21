package courses

import (
	"context"

	"cu-sync/internal/cu"
)

// LMSClient defines the subset of the LMS API needed by the courses usecase.
type LMSClient interface {
	GetStudentCourses(ctx context.Context, limit int, state string) (*cu.StudentCoursesResponse, error)
}
