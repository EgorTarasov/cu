package task

import (
	"context"
	"cu-sync/internal/gateway/cu"
)

// LMSClient defines the subset of the LMS API needed by the task usecase.
type LMSClient interface {
	GetTask(ctx context.Context, taskID int) (*cu.Task, error)
}
