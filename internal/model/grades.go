package model

// GradesSummaryInput is the input for the grades summary.
type GradesSummaryInput struct{}

// GradesSummaryItem represents one course in the grades summary.
type GradesSummaryItem struct {
	CourseName  string
	EarnedScore float64
	MaxScore    float64
	Error       error
}

// GradesSummaryOutput is the result of the grades summary.
type GradesSummaryOutput struct {
	Items []GradesSummaryItem
}

// GradesDetailedInput is the input for detailed grades of a specific course.
type GradesDetailedInput struct {
	CourseQuery string
}

// ActivityBreakdown represents one activity's performance.
type ActivityBreakdown struct {
	Name      string
	Weight    float64
	Average   float64
	Total     float64
	IsBlocker bool
}

// TaskGrade represents a single task's grade.
type TaskGrade struct {
	Name       string
	State      string
	StateLabel string
	Score      *float64
	MaxScore   int
}

// BlockerInfo describes a blocker activity.
type BlockerInfo struct {
	ActivityName string
	Threshold    float64
}

// GradesDetailedOutput is the result of detailed grades for a course.
type GradesDetailedOutput struct {
	CourseName string
	Activities []ActivityBreakdown
	TotalScore float64
	Tasks      []TaskGrade
	Blockers   []BlockerInfo
}
