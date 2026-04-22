package model

type GradesSummaryInput struct{}

type GradesSummaryItem struct {
	CourseName  string
	EarnedScore float64
	MaxScore    float64
	Error       error
}

type GradesSummaryOutput struct {
	Items []GradesSummaryItem
}

type GradesDetailedInput struct {
	CourseQuery string
}

type ActivityBreakdown struct {
	Name      string
	Weight    float64
	Average   float64
	Total     float64
	IsBlocker bool
}

type TaskGrade struct {
	Name     string
	State    TaskState
	Score    *float64
	MaxScore int
}

type BlockerInfo struct {
	ActivityName string
	Threshold    float64
}

type GradesDetailedOutput struct {
	CourseName string
	Activities []ActivityBreakdown
	TotalScore float64
	Tasks      []TaskGrade
	Blockers   []BlockerInfo
}
