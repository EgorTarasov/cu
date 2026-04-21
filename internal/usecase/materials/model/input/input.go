package input

// DownloadInput is the input for downloading course materials.
type DownloadInput struct {
	CourseQuery string
	WeekFilter  int
	LinksOnly   bool
	BasePath    string
}
