package model

type CoursesListInput struct{}

type CourseItem struct {
	ID         int
	Name       string
	IsArchived bool
}

type CoursesListOutput struct {
	Active   []CourseItem
	Archived []CourseItem
}
