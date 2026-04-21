package output

import "time"

// UrgencyLevel classifies how urgent a deadline is.
type UrgencyLevel string

const (
	UrgencyUrgent UrgencyLevel = "urgent"
	UrgencySoon   UrgencyLevel = "soon"
	UrgencyNormal UrgencyLevel = "normal"
)

// DeadlineItem represents a single deadline in the output.
type DeadlineItem struct {
	ExerciseName string
	CourseName   string
	State        string
	StateLabel   string
	Deadline     time.Time
	TimeLeft     string
	Urgency      UrgencyLevel
	ReviewerName string
}

// ListOutput is the result of listing deadlines.
type ListOutput struct {
	Items      []DeadlineItem
	CourseName string
}
