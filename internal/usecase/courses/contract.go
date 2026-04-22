package courses

import (
	"context"
	"cu-sync/internal/gateway/cu"
)

type LMSClient interface {
	GetStudentCourses(ctx context.Context, limit int, state string) (*cu.StudentCoursesResponse, error)
}
