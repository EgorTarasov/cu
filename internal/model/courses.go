package model

type CoursesListInput struct{}

type CourseItem struct {
	ID   int
	Name string
}

type CoursesListOutput struct {
	Items []CourseItem
}
