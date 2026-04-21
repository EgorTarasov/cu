package model

// MaterialsDownloadInput is the input for downloading course materials.
type MaterialsDownloadInput struct {
	CourseQuery string
	WeekFilter  int
	LinksOnly   bool
	BasePath    string
}

// MaterialEventType identifies the kind of material event.
type MaterialEventType string

const (
	MaterialEventTheme MaterialEventType = "theme"
	MaterialEventPDF   MaterialEventType = "pdf"
	MaterialEventLink  MaterialEventType = "link"
	MaterialEventSaved MaterialEventType = "saved"
	MaterialEventError MaterialEventType = "error"
)

// MaterialEvent represents a single event during material download.
type MaterialEvent struct {
	Type     MaterialEventType
	Message  string
	FilePath string
}

// MaterialsDownloadOutput summarises the download results.
type MaterialsDownloadOutput struct {
	TotalFiles      int32
	DownloadedFiles int32
}
