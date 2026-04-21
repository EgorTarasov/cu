package output

import "time"

// TaskOutput contains the task details with computed fields.
type TaskOutput struct {
	// Direct fields from the task.
	CourseName      string
	ThemeName       string
	ExerciseName    string
	ActivityName    string
	ActivityWeight  float64
	Deadline        time.Time
	StartedAt       *time.Time
	SubmitAt        *time.Time
	RejectAt        *time.Time
	EvaluateAt      *time.Time
	ReviewerName    string
	ReviewerEmail   string
	SolutionURL     string
	MaxScore        int
	LateDaysBalance int

	// Computed fields.
	StateLabel     string
	TimeLeft       string
	ScoreFormatted string
}
