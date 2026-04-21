package input

import "time"

// LoginInput is the input for authentication.
type LoginInput struct {
	Timeout time.Duration
}
