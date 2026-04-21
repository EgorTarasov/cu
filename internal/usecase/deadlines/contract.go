package deadlines

import (
	"context"

	"cu-sync/internal/cu"
)

// LMSClient defines the subset of the LMS API needed by the deadlines usecase.
type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetDeadlines(ctx context.Context, limit int, courseID *int) ([]cu.Deadline, error)
}
