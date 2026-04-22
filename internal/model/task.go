package model

import (
	"strings"
	"time"
)

type TaskState string

const (
	TaskBacklog    TaskState = "backlog"
	TaskInProgress TaskState = "inProgress"
	TaskSubmitted  TaskState = "submitted"
	TaskEvaluated  TaskState = "evaluated"
	TaskFailed     TaskState = "failed"
)

func (t TaskState) Label() string {
	switch t {
	case TaskBacklog:
		return "TODO"
	case TaskInProgress:
		return "IN PROGRESS"
	case TaskSubmitted:
		return "SUBMITTED"
	case TaskEvaluated:
		return "DONE"
	case TaskFailed:
		return "FAILED"
	default:
		return strings.ToUpper(string(t))
	}
}

type Reviewer struct {
	FirstName string
	LastName  string
	Email     string
}

func (r Reviewer) FullName() string {
	return r.FirstName + " " + r.LastName
}

type TaskGetInput struct {
	TaskID int
}

type TaskOutput struct {
	CourseName      string
	ThemeName       string
	ExerciseName    string
	ActivityName    string
	ActivityWeight  float64
	Deadline        DeadLine
	StartedAt       *time.Time
	SubmitAt        *time.Time
	RejectAt        *time.Time
	EvaluateAt      *time.Time
	Reviewer        *Reviewer
	SolutionURL     string
	MaxScore        int
	LateDaysBalance int

	StateLabel     string
	ScoreFormatted string
}
