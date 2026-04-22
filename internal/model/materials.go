package model

type MaterialsDownloadInput struct {
	CourseQuery string
	WeekFilter  int
	LinksOnly   bool
	BasePath    string
}

type MaterialEventType string

const (
	MaterialEventTheme MaterialEventType = "theme"
	MaterialEventPDF   MaterialEventType = "pdf"
	MaterialEventLink  MaterialEventType = "link"
	MaterialEventSaved MaterialEventType = "saved"
	MaterialEventError MaterialEventType = "error"
)

type MaterialEvent struct {
	Type     MaterialEventType
	Message  string
	FilePath string
}

type MaterialsDownloadOutput struct {
	TotalFiles      int32
	DownloadedFiles int32
}
