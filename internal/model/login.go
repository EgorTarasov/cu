package model

import "time"

type LoginInput struct {
	Timeout time.Duration
}

type LoginOutput struct {
	CookiePath      string
	ValidationError error
}
