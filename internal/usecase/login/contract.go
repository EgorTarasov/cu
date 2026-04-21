package login

import (
	"context"
	"time"
)

// AuthFunc performs browser-based authentication and returns a cookie.
type AuthFunc func(ctx context.Context, timeout time.Duration) (cookie string, err error)

// SaveFunc persists a cookie string.
type SaveFunc func(cookie string) error

// ValidateFunc validates a cookie. May be nil.
type ValidateFunc func() error
