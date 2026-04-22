package deadlines

import (
	"context"
	"cu-sync/internal/gateway/cu"
)

type LMSClient interface {
	ResolveCourse(ctx context.Context, query string) (int, string, error)
	GetDeadlines(ctx context.Context, limit int, courseID *int) ([]cu.Deadline, error)
}
