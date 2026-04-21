package output

// EventType identifies the kind of material event.
type EventType string

const (
	EventTheme EventType = "theme"
	EventPDF   EventType = "pdf"
	EventLink  EventType = "link"
	EventSaved EventType = "saved"
	EventError EventType = "error"
)

// MaterialEvent represents a single event during material download.
type MaterialEvent struct {
	Type     EventType
	Message  string
	FilePath string
}

// DownloadOutput summarises the download results.
type DownloadOutput struct {
	TotalFiles      int32
	DownloadedFiles int32
}
