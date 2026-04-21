package model

// CoursesListInput is the input for listing courses.
type CoursesListInput struct{}

// CourseItem represents a single course in the list.
type CourseItem struct {
	ID   int
	Name string
}

// CoursesListOutput is the result of listing courses.
type CoursesListOutput struct {
	Items []CourseItem
}
