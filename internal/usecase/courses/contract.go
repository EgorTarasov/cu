package courses

import (
	"context"

	"github.com/EgorTarasov/cu/internal/gateway/cu"
)

type LMSClient interface {
	GetStudentCourses(ctx context.Context, limit int, state string) (*cu.StudentCoursesResponse, error)
}
