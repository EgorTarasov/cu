package model

import "time"

// LoginInput is the input for authentication.
type LoginInput struct {
	Timeout time.Duration
}

// LoginOutput is the result of an authentication flow.
type LoginOutput struct {
	CookiePath      string
	ValidationError error
}
