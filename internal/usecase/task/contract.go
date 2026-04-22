package task

import (
	"context"
	"cu-sync/internal/gateway/cu"
)

type LMSClient interface {
	GetTask(ctx context.Context, taskID int) (*cu.Task, error)
}
