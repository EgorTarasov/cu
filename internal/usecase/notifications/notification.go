package notifications

import (
	"context"
	"time"
)

// Kind enumerates the bot post headers we recognize.
const (
	KindNewTask   = "task_new"
	KindGraded    = "task_graded"
	KindRecording = "recording"
	KindOther     = "other"
)

// Notification is a structured view of one bot message linking back to LMS content.
type Notification struct {
	PostID     string
	PostURL    string // mm permalink, if available
	PostedAt   time.Time
	Kind       string
	Title      string // link title or first non-header line
	Excerpt    string // short snippet of the message
	CourseID   int    // 0 if not parseable
	ThemeID    int
	LongreadID int
	LMSURL     string // the linked LMS URL, if any
}

// RawPost mirrors the shape used by other usecases.
type RawPost struct {
	ID       string
	Message  string
	CreateAt time.Time
}

// PostSource hands back the post stream to scan.
type PostSource interface {
	LoadPosts(ctx context.Context) ([]RawPost, error)
}

// Permalinker turns a post ID into a chat permalink. Optional.
type Permalinker interface {
	Permalink(postID string) string
}
