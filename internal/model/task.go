package model

import "time"

// TaskGetInput is the input for fetching a single task.
type TaskGetInput struct {
	TaskID int
}

// TaskOutput contains the task details with computed fields.
type TaskOutput struct {
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

	StateLabel     string
	TimeLeft       string
	ScoreFormatted string
}
