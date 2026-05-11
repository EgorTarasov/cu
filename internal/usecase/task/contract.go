package task

import (
	"context"

	"github.com/EgorTarasov/cu/internal/gateway/cu"
)

type LMSClient interface {
	GetTask(ctx context.Context, taskID int) (*cu.Task, error)
}
