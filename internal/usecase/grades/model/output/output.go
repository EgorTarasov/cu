package output

// SummaryItem represents one course in the grades summary.
type SummaryItem struct {
	CourseName  string
	EarnedScore float64
	MaxScore    float64
	Error       error
}

// SummaryOutput is the result of the grades summary.
type SummaryOutput struct {
	Items []SummaryItem
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

// DetailedOutput is the result of detailed grades for a course.
type DetailedOutput struct {
	CourseName string
	Activities []ActivityBreakdown
	TotalScore float64
	Tasks      []TaskGrade
	Blockers   []BlockerInfo
}
