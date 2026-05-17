package recordings

import (
	"context"
	"time"
)

// RawPost is the minimal shape the usecase needs from any post source.
type RawPost struct {
	ID       string
	Message  string
	CreateAt time.Time
}

// Permalinker turns a post ID into a chat permalink. Optional — usecase works
// without it, recordings just won't have PostURL filled.
type Permalinker interface {
	Permalink(postID string) string
}

// PostSource is implemented by any store that can hand back the posts to scan.
// The implementation is free to read from local cache or remote API.
type PostSource interface {
	LoadPosts(ctx context.Context) ([]RawPost, error)
}
