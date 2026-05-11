package cu

import "context"

// GetCurrentStudent fetches the authenticated student profile.
func (c *Client) GetCurrentStudent(ctx context.Context) (*Student, error) {
	return doJSON[Student](ctx, c, CurrentStudentEndpoint)
}
