package output

// CourseItem represents a single course in the list.
type CourseItem struct {
	ID   int
	Name string
}

// ListOutput is the result of listing courses.
type ListOutput struct {
	Items []CourseItem
}
