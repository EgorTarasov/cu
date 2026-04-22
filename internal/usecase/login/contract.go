package login

import (
	"context"
	"time"
)

type AuthFunc func(ctx context.Context, timeout time.Duration) (cookie string, err error)

type SaveFunc func(cookie string) error

type ValidateFunc func() error
