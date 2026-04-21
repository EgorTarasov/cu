package input

// SummaryInput is the input for the grades summary.
type SummaryInput struct{}

// DetailedInput is the input for detailed grades of a specific course.
type DetailedInput struct {
	CourseQuery string
}
