package output

// LoginOutput is the result of an authentication flow.
type LoginOutput struct {
	CookiePath      string
	ValidationError error
}
