package model

import "time"

// DeadlinesListInput is the input for listing deadlines.
type DeadlinesListInput struct {
	CourseQuery string
}

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

// DeadlinesListOutput is the result of listing deadlines.
type DeadlinesListOutput struct {
	Items      []DeadlineItem
	CourseName string
}
